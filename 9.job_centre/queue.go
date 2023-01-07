package main

import (
	"encoding/json"
	"sync"
)

type Queue struct {
	m           *sync.Mutex
	jobs        map[int]*job
	idGenerator *jobIDGenerator
}

func newQueue(idGenerator *jobIDGenerator) *Queue {
	return &Queue{
		m:           &sync.Mutex{},
		idGenerator: idGenerator,
		jobs:        make(map[int]*job),
	}
}

func (t *Queue) delete(jobID int) bool {
	t.m.Lock()
	defer t.m.Unlock()

	_, found := t.jobs[jobID]
	if !found {
		return false
	}

	delete(t.jobs, jobID)
	return true
}

func (t *Queue) abort(jobID int) bool {
	t.m.Lock()
	defer t.m.Unlock()

	job, found := t.jobs[jobID]
	if !found {
		return false
	}

	job.isBeingWorkedOn = false
	t.jobs[jobID] = job

	return true
}

func (t *Queue) put(queue string, priority int, content json.RawMessage) job {
	t.m.Lock()
	defer t.m.Unlock()

	id := t.idGenerator.gen()
	j := job{
		id:       id,
		priority: priority,
		content:  content,
	}
	t.jobs[id] = &j
	return j
}

func (t *Queue) getHighestPriorityJob() *job {
	t.m.Lock()
	defer t.m.Unlock()

	if len(t.jobs) == 0 {
		return nil
	}

	var highestPriorityJob *job
	for _, job := range t.jobs {
		if highestPriorityJob == nil || job.priority >= highestPriorityJob.priority {
			highestPriorityJob = job
		}
	}

	return highestPriorityJob
}
