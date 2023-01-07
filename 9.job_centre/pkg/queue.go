package pkg

import (
	"sync"
)

type Queue struct {
	name string
	m    *sync.Mutex
	jobs map[int]*Job
}

func newQueue(name string) *Queue {
	return &Queue{
		name: name,
		m:    &sync.Mutex{},
		jobs: make(map[int]*Job),
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

func (t *Queue) putJob(j *Job) {
	t.m.Lock()
	defer t.m.Unlock()

	t.jobs[j.ID] = j
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
