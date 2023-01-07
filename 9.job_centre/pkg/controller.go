package pkg

import (
	"encoding/json"
	"sync"
)

type JobController struct {
	m           *sync.Mutex
	queues      map[string]*Queue
	idGenerator *jobIDGenerator
	activeJobs  *activeJobsController
	locker      *lock
}

func NewJobController() *JobController {
	return &JobController{
		m:           &sync.Mutex{},
		queues:      make(map[string]*Queue),
		idGenerator: newJobIDGenerator(),
		activeJobs:  newActiveJobsController(),
		locker:      newLock(),
	}
}

func (t *JobController) Leave(clientID int) {
	t.locker.leave(clientID)
}

func (t *JobController) Put(qName string, priority int, content json.RawMessage) Job {
	t.m.Lock()
	q, ok := t.queues[qName]
	if !ok {
		q = newQueue(t.idGenerator)
		t.queues[qName] = q
	}
	t.m.Unlock()

	job := q.put(qName, priority, content)
	t.locker.announce()
	return job
}

func (t *JobController) GetWithWait(clientID int, qNames []string, wait bool) (*Job, string) {
	job, qName := t.get(clientID, qNames)
	if job != nil {
		return job, qName
	}

	if !wait {
		return nil, ""
	}

	for {
		t.locker.wait(clientID)
		job, qName = t.get(clientID, qNames)
		if job == nil {
			continue // keep waiting
		}

		return job, qName
	}
}

func (t *JobController) IsWaiting(clientID int) bool {
	return t.locker.isWaiting(clientID)
}

func (t *JobController) Abort(clientID, jobID int) bool {
	releasedJob := t.activeJobs.release(clientID, jobID)
	if releasedJob == nil {
		return false
	}

	t.m.Lock()
	t.queues[releasedJob.qName].putWithID(releasedJob.qName, releasedJob.job.ID, releasedJob.job.Priority, releasedJob.job.Content)
	t.locker.announce()
	t.m.Unlock()
	return true
}

func (t *JobController) Delete(clientID, jobID int) bool {
	t.m.Lock()
	defer t.m.Unlock()

	for i := range t.queues {
		if t.queues[i].delete(jobID) {
			return true
		}
	}

	// Maybe, it's in active jobs?
	releasedJob := t.activeJobs.release(clientID, jobID)
	return releasedJob != nil
}

func (t *JobController) ReleaseActiveJobs(clientID int) {
	releasedJobs := t.activeJobs.releaseAll(clientID)

	t.m.Lock()
	for _, aj := range releasedJobs {
		t.queues[aj.qName].putWithID(aj.qName, aj.job.ID, aj.job.Priority, aj.job.Content)
		t.locker.announce()
	}
	t.m.Unlock()
}

func (t *JobController) addToActiveJobs(clientID int, qName string, queue *Queue, job *Job) {
	aj := &activeJob{qName: qName, job: job}
	t.activeJobs.add(clientID, aj)
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

		if highestPriorityJob == nil || h.Priority >= highestPriorityJob.Priority {
			highestPriorityJob = h
			correspondingQueue = q
			correspondingQueueName = qName
		}
	}

	// If found, then remove the job from the queue
	// And add it to the active jobs list
	if highestPriorityJob != nil {
		t.addToActiveJobs(clientID, correspondingQueueName, correspondingQueue, highestPriorityJob)
	}

	return highestPriorityJob, correspondingQueueName
}
