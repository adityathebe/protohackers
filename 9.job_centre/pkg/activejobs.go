package pkg

import "sync"

type activeJob struct {
	qName string
	job   *Job
}

type activeJobsController struct {
	m          *sync.Mutex
	activeJobs map[int][]*activeJob
}

func newActiveJobsController() *activeJobsController {
	return &activeJobsController{
		m:          &sync.Mutex{},
		activeJobs: make(map[int][]*activeJob),
	}
}

func (t *activeJobsController) releaseAll(clientID int) []*activeJob {
	t.m.Lock()
	defer t.m.Unlock()

	activeJobs := t.activeJobs[clientID]
	delete(t.activeJobs, clientID)

	return activeJobs
}

func (t *activeJobsController) release(clientID, jobID int) *activeJob {
	t.m.Lock()
	defer t.m.Unlock()

	activeJobs := t.activeJobs[clientID]
	for _, aj := range activeJobs {
		if aj.job.ID == jobID {
			delete(t.activeJobs, clientID)
			return aj
		}
	}

	return nil
}

func (t *activeJobsController) add(clientID int, aj *activeJob) {
	t.m.Lock()
	t.activeJobs[clientID] = append(t.activeJobs[clientID], aj)
	t.m.Unlock()
}
