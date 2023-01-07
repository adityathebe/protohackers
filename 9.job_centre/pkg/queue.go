package pkg

import (
	"encoding/json"
	"sync"
)

type Queue struct {
	m           *sync.Mutex
	jobs        map[int]*Job
	idGenerator *jobIDGenerator
}

func newQueue(idGenerator *jobIDGenerator) *Queue {
	return &Queue{
		m:           &sync.Mutex{},
		idGenerator: idGenerator,
		jobs:        make(map[int]*Job),
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

func (t *Queue) getByID(jobID int) *Job {
	t.m.Lock()
	defer t.m.Unlock()

	return t.jobs[jobID]
}

func (t *Queue) put(queue string, priority int, content json.RawMessage) Job {
	id := t.idGenerator.gen()
	j := t.putWithID(queue, id, priority, content)
	return j
}

func (t *Queue) putWithID(queue string, id, priority int, content json.RawMessage) Job {
	t.m.Lock()
	defer t.m.Unlock()

	j := Job{
		ID:       id,
		Priority: priority,
		Content:  content,
	}
	t.jobs[id] = &j
	return j
}

func (t *Queue) getHighestPriorityJob() *Job {
	t.m.Lock()
	defer t.m.Unlock()

	if len(t.jobs) == 0 {
		return nil
	}

	var highestPriorityJob *Job
	for _, job := range t.jobs {
		if highestPriorityJob == nil || job.Priority >= highestPriorityJob.Priority {
			highestPriorityJob = job
		}
	}

	return highestPriorityJob
}
