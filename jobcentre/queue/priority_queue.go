package queue

import "container/heap"

// PriorityQueue store all the jobs in a particular queue
type PriorityQueue []*Job

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].heapPriority > pq[j].heapPriority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*Job)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Peek() *Job {
	return (*pq)[0]
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) Update(item *Job, assignedTo *int, heapPriority int) {
	item.assignedTo = assignedTo
	item.heapPriority = heapPriority
	heap.Fix(pq, item.index)
}
