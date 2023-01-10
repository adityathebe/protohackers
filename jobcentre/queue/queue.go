package queue

import (
	"container/heap"
	"encoding/json"
	"errors"
	"sync"
)

type Queue struct {
	idGenerator  *generator
	m            *sync.Mutex
	queues       map[string]*PriorityQueue
	jobMap       map[int]*Job
	subscription *subscription
}

func NewQueue() *Queue {
	return &Queue{
		m:            &sync.Mutex{},
		queues:       make(map[string]*PriorityQueue, 2000),
		jobMap:       make(map[int]*Job, 2000),
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

func (t *Queue) Put(qName string, priority int, content json.RawMessage) *Job {
	t.m.Lock()
	defer t.m.Unlock()

	job := &Job{
		ID:           t.idGenerator.Gen(),
		Priority:     priority,
		heapPriority: priority,
		Content:      content,
		QName:        qName,
	}

	pq, ok := t.queues[qName]
	if !ok {
		pqs := make(PriorityQueue, 0, 100)
		heap.Init(&pqs)
		pq = &pqs
	}

	heap.Push(pq, job)
	t.queues[qName] = pq
	t.jobMap[job.ID] = job

	t.announce()
	return job
}

func (t *Queue) Abort(clientID, jobID int, shouldAuthorize bool) (aborted bool, error error) {
	t.m.Lock()
	defer t.m.Unlock()

	r, ok := t.jobMap[jobID]
	if !ok {
		return false, nil
	}

	if r.assignedTo == nil {
		return false, nil
	}

	if shouldAuthorize && *r.assignedTo != clientID {
		return false, errors.New("unauthorized_to_abort")
	}

	queue := t.queues[r.QName]
	queue.Update(r, nil, -r.Priority)
	t.announce()
	return true, nil
}

func (t *Queue) Delete(jobID int) bool {
	t.m.Lock()
	defer t.m.Unlock()

	job, ok := t.jobMap[jobID]
	if !ok {
		return false
	}

	pq := t.queues[job.QName]
	heap.Remove(pq, job.index)
	delete(t.jobMap, jobID)
	return true
}

func (t *Queue) ReleaseActiveJobs(clientID int) {
	t.m.Lock()
	defer t.m.Unlock()

	for qName, records := range t.queues {
		queue := t.queues[qName]
		for _, job := range *records {
			if job.assignedTo != nil && *job.assignedTo == clientID {
				queue.Update(job, nil, -job.Priority)
			}
		}
	}

	t.announce()
}

func (t *Queue) Get(clientID int, qNames []string) *Job {
	t.m.Lock()
	defer t.m.Unlock()

	var response *Job
	for _, qName := range qNames {
		queue := t.queues[qName]
		if queue == nil || queue.Len() == 0 {
			continue
		}

		j := queue.Peek()
		if j.assignedTo != nil {
			continue
		}

		if response == nil || j.Priority >= response.Priority {
			response = j
		}
	}

	if response == nil {
		return nil
	}

	// If found, then remove the job from the queue
	// And add it to the active jobs list
	queue := t.queues[response.QName]
	queue.Update(response, &clientID, -response.Priority)
	return response
}
