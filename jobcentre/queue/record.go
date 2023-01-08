package queue

import (
	jobcentre "github.com/adityathebe/protohackers/jobcentre"
)

type Record struct {
	Job        *jobcentre.Job
	QName      string
	assignedTo *int
}

func newRecord(qName string, job *jobcentre.Job) *Record {
	return &Record{
		QName: qName,
		Job:   job,
	}
}
