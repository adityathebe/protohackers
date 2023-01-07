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

func (t *JobController) Put(qName string, priority int, content json.RawMessage) *Job {
	t.m.Lock()
	q, ok := t.queues[qName]
	if !ok {
		q = newQueue(qName)
		t.queues[qName] = q
	}
	t.m.Unlock()

	job := &Job{
		ID:       t.idGenerator.gen(),
		Priority: priority,
		Content:  content,
	}
	q.putJob(job)
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

func (t *JobController) Abort(clientID, jobID int) (bool, bool) {
	releasedJob, unauthorized := t.activeJobs.release(clientID, jobID, true)
	if unauthorized {
		return false, true
	}

	if releasedJob == nil {
		return false, false
	}

	t.m.Lock()
	t.queues[releasedJob.qName].putJob(releasedJob.job)
	t.locker.announce()
	t.m.Unlock()
	return true, false
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
	releasedJob, unauthorized := t.activeJobs.release(clientID, jobID, false)
	if unauthorized {
		panic("should not happen")
	}
	return releasedJob != nil
}

func (t *JobController) ReleaseActiveJobs(clientID int) {
	releasedJobs := t.activeJobs.releaseAll(clientID)

	t.m.Lock()
	for _, aj := range releasedJobs {
		t.queues[aj.qName].putJob(aj.job)
	}
	t.locker.announce()
	t.m.Unlock()
}

func (t *JobController) get(clientID int, qNames []string) (*Job, string) {
	var (
		highestPriorityJob *Job
		correspondingQueue *Queue
	)

	t.m.Lock()
	for _, qName := range qNames {
		q, ok := t.queues[qName]
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
		}
	}
	t.m.Unlock()

	if highestPriorityJob == nil {
		return nil, ""
	}

	// If found, then remove the job from the queue
	// And add it to the active jobs list
	aj := &activeJob{qName: correspondingQueue.name, job: highestPriorityJob}
	t.activeJobs.add(clientID, aj)
	correspondingQueue.delete(highestPriorityJob.ID)

	return highestPriorityJob, correspondingQueue.name
}
