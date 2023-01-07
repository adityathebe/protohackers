package main

import (
	"encoding/json"
	"sync"
)

type jobQueue struct {
	m           *sync.Mutex
	queues      map[string]*Queue
	idGenerator *jobIDGenerator
}

func newJobQueue() *jobQueue {
	return &jobQueue{
		m:           &sync.Mutex{},
		queues:      make(map[string]*Queue),
		idGenerator: newJobIDGenerator(),
	}
}

func (t *jobQueue) put(qName string, priority int, content json.RawMessage) job {
	t.m.Lock()
	defer t.m.Unlock()

	q, ok := t.queues[qName]
	if !ok {
		q = newQueue(t.idGenerator)
		t.queues[qName] = q
	}

	return q.put(qName, priority, content)
}

func (t *jobQueue) get(queues []string, wait bool) (*job, string) {
	t.m.Lock()
	defer t.m.Unlock()

	var highestPriorityJob *job
	var correspondingQueue string
	for _, qName := range queues {
		q, ok := t.queues[qName]
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

func (t *jobQueue) abort(jobID int) bool {
	for i := range t.queues {
		if t.queues[i].abort(jobID) {
			return true
		}
	}

	return false
}

func (t *jobQueue) delete(jobID int) bool {
	for i := range t.queues {
		if t.queues[i].delete(jobID) {
			return true
		}
	}

	return false
}
