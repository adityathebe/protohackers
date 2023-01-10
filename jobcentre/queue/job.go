package queue

import "encoding/json"

type Job struct {
	ID         int
	Content    json.RawMessage
	Priority   int
	QName      string
	assignedTo *int

	// index is the position in heap
	index int

	heapPriority int
}
