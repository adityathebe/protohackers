package queue

import (
	"encoding/json"
	"errors"
	"sync"

	jobcentre "github.com/adityathebe/protohackers/jobcentre"
)

type Queue struct {
	idGenerator  *generator
	m            *sync.Mutex
	records      []*Record
	subscription *subscription
}

func NewQueue() *Queue {
	return &Queue{
		m:            &sync.Mutex{},
		idGenerator:  newIDGenerator(),
		subscription: newSubscription(),
	}
}

func (t *Queue) Subscribe(id int) <-chan struct{} {
	return t.subscription.Subscribe(id)
}

func (t *Queue) Unsubscribe(id int) {
	t.subscription.Unsubscribe(id)
}

func (t *Queue) announce() {
	t.subscription.announce()
}

func (t *Queue) Put(qName string, priority int, content json.RawMessage) *jobcentre.Job {
	job := &jobcentre.Job{
		ID:       t.idGenerator.Gen(),
		Priority: priority,
		Content:  content,
	}

	record := newRecord(qName, job)
	t.m.Lock()
	t.records = append(t.records, record)
	t.m.Unlock()

	t.announce()
	return job
}

func (t *Queue) Abort(clientID, jobID int, shouldAuthorize bool) (aborted bool, error error) {
	t.m.Lock()
	defer t.m.Unlock()

	for i, r := range t.records {
		if r.Job.ID == jobID {
			if r.assignedTo == nil {
				return false, nil
			}

			if shouldAuthorize && *r.assignedTo != clientID {
				return false, errors.New("unauthorized_to_abort")
			}

			t.records[i].assignedTo = nil
			t.announce()
			return true, nil
		}
	}

	return false, nil
}

func (t *Queue) Delete(jobID int) bool {
	t.m.Lock()
	defer t.m.Unlock()

	for i, r := range t.records {
		if r.Job.ID == jobID {
			t.records = append(t.records[:i], t.records[i+1:]...)
			return true
		}
	}

	return false
}

func (t *Queue) ReleaseActiveJobs(clientID int) {
	t.m.Lock()
	defer t.m.Unlock()

	for i, r := range t.records {
		if r.assignedTo != nil && *r.assignedTo == clientID {
			t.records[i].assignedTo = nil
		}
	}

	t.announce()
}

func (t *Queue) Get(clientID int, qNames []string) *Record {
	t.m.Lock()
	defer t.m.Unlock()

	var response *Record
	for _, r := range t.records {
		if !isInSlice(qNames, r.QName) {
			continue
		}

		if (response == nil || r.Job.Priority >= response.Job.Priority) && r.assignedTo == nil {
			response = r
		}
	}

	if response == nil {
		return nil
	}

	// If found, then remove the job from the queue
	// And add it to the active jobs list
	response.assignedTo = &clientID
	return response
}

func isInSlice(s []string, item string) bool {
	for _, i := range s {
		if item == i {
			return true
		}
	}

	return false
}
