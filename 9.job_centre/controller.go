package main

import (
	"encoding/json"
	"sync"
	"time"
)

type JobController struct {
	m           *sync.Mutex
	queues      map[string]*Queue
	idGenerator *jobIDGenerator
}

func newJobController() *JobController {
	return &JobController{
		m:           &sync.Mutex{},
		queues:      make(map[string]*Queue),
		idGenerator: newJobIDGenerator(),
	}
}

func (t *JobController) put(qName string, priority int, content json.RawMessage) Job {
	t.m.Lock()
	q, ok := t.queues[qName]
	if !ok {
		q = newQueue(t.idGenerator)
		t.queues[qName] = q
	}
	t.m.Unlock()

	return q.put(qName, priority, content)
}

func (t *JobController) getWithWait(qNames []string, wait bool) (*Job, string) {
	job, qName := t.get(qNames)
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
		job, qName := t.get(qNames)
		if job != nil {
			return job, qName
		}
	}
}

func (t *JobController) get(qNames []string) (*Job, string) {
	var highestPriorityJob *Job
	var correspondingQueue string
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

		if highestPriorityJob == nil || highestPriorityJob.priority >= h.priority {
			highestPriorityJob = h
			correspondingQueue = qName
		}
	}

	return highestPriorityJob, correspondingQueue
}

func (t *JobController) abort(jobID int) bool {
	for i := range t.queues {
		if t.queues[i].abort(jobID) {
			return true
		}
	}

	return false
}

func (t *JobController) delete(jobID int) bool {
	for i := range t.queues {
		if t.queues[i].delete(jobID) {
			return true
		}
	}

	return false
}
