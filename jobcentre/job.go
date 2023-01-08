package jobcentre

import (
	"encoding/json"
)

type Job struct {
	ID       int
	Content  json.RawMessage
	Priority int
}
