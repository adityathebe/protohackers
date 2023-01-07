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

func (t *activeJobsController) release(requesterID, jobID int, authorizedOnly bool) (*activeJob, bool) {
	t.m.Lock()
	defer t.m.Unlock()

	for clientID, activeJobs := range t.activeJobs {
		for _, aj := range activeJobs {
			if aj.job.ID == jobID {
				if authorizedOnly && requesterID != clientID {
					return nil, true
				}

				delete(t.activeJobs, clientID)
				return aj, false
			}
		}
	}

	return nil, false
}

func (t *activeJobsController) add(clientID int, aj *activeJob) {
	t.m.Lock()
	t.activeJobs[clientID] = append(t.activeJobs[clientID], aj)
	t.m.Unlock()
}
