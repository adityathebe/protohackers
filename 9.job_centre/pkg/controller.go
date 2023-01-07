package pkg

import (
	"encoding/json"
	"log"
	"sync"
	"time"
)

type JobController struct {
	m           *sync.Mutex
	queues      map[string]*Queue
	idGenerator *jobIDGenerator
	activeJobs  map[int][]*ActiveJob
}

type ActiveJob struct {
	qName string
	job   *Job
}

func NewJobController() *JobController {
	return &JobController{
		m:           &sync.Mutex{},
		queues:      make(map[string]*Queue),
		idGenerator: newJobIDGenerator(),
		activeJobs:  make(map[int][]*ActiveJob),
	}
}

func (t *JobController) Put(qName string, priority int, content json.RawMessage) Job {
	t.m.Lock()
	q, ok := t.queues[qName]
	if !ok {
		q = newQueue(t.idGenerator)
		t.queues[qName] = q
	}
	t.m.Unlock()

	return q.put(qName, priority, content)
}

func (t *JobController) GetWithWait(clientID int, qNames []string, wait bool) (*Job, string) {
	job, qName := t.get(clientID, qNames)
	if job != nil {
		return job, qName
	}

	if !wait {
		return nil, ""
	}

	// Instead of waiting for a second everytime,
	// I could probably use channels to signal whenever something is put
	// into the job queue.
	for {
		time.Sleep(time.Second)
		job, qName := t.get(clientID, qNames)
		if job != nil {
			return job, qName
		}
	}
}

func (t *JobController) Abort(clientID, jobID int) bool {
	for qName := range t.queues {
		job := t.queues[qName].getByID(jobID)
		if job == nil {
			continue
		}

		t.activateJob(clientID, qName, t.queues[qName], job)
	}

	return false
}

func (t *JobController) Delete(jobID int) bool {
	for i := range t.queues {
		if t.queues[i].delete(jobID) {
			return true
		}
	}

	return false
}

func (t *JobController) ReleaseActiveJobs(clientID int) {
	log.Printf("Releasing [%d] active jobs of client [%d]", len(t.activeJobs[clientID]), clientID)
	for _, aj := range t.activeJobs[clientID] {
		t.queues[aj.qName].putWithID(aj.qName, aj.job.ID, aj.job.Priority, aj.job.Content)
	}
}

func (t *JobController) activateJob(clientID int, qName string, queue *Queue, job *Job) {
	aj := &ActiveJob{qName: qName, job: job}
	t.activeJobs[clientID] = append(t.activeJobs[clientID], aj)
	queue.delete(job.ID)
}

func (t *JobController) get(clientID int, qNames []string) (*Job, string) {
	var highestPriorityJob *Job
	var correspondingQueue *Queue
	var correspondingQueueName string
	for _, qName := range qNames {
		t.m.Lock()
		q, ok := t.queues[qName]
		t.m.Unlock()
		if !ok {
			continue
		}

		h := q.getHighestPriorityJob()
		if h == nil {
			continue
		}

		if highestPriorityJob == nil || highestPriorityJob.Priority >= h.Priority {
			highestPriorityJob = h
			correspondingQueue = q
			correspondingQueueName = qName
		}
	}

	// If found, then remove the job from the queue
	// And add it to the active jobs list
	if highestPriorityJob != nil {
		t.activateJob(clientID, correspondingQueueName, correspondingQueue, highestPriorityJob)
	}

	return highestPriorityJob, correspondingQueueName
}
