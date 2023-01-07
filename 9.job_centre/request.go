package main

import "encoding/json"

type Request struct {
	Request  string          `json:"request,omitempty"` // "put", "get", "delet", "abort"
	Queue    string          `json:"queue,omitempty"`
	Queues   []string        `json:"queues,omitempty"`
	Priority *int            `json:"pri,omitempty"`
	Job      json.RawMessage `json:"job,omitempty"`
	Wait     bool            `json:"wait,omitempty"`
	ID       *int            `json:"id,omitempty"`
}

func (t *Request) isValid() bool {
	switch t.Request {
	case "put":
		if t.Priority == nil {
			return false
		}

		if t.Queue == "" {
			return false
		}

	case "get":
		if len(t.Queues) == 0 {
			return false
		}

	case "delete", "abort":
		if t.ID == nil {
			return false
		}

	default:
		return false
	}

	return true
}

type Response struct {
	Status   string          `json:"status"` // "ok", "error", or "no-job".
	ID       int             `json:"id,omitempty"`
	Queue    string          `json:"queue,omitempty"`
	Priority int             `json:"pri,omitempty"`
	Job      json.RawMessage `json:"job,omitempty"`
}