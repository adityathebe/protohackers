package client

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type logger struct {
	f *os.File
}

func newLog(clientID int) *logger {
	filename := fmt.Sprintf("jobcentre/logs/%d.txt", clientID)
	f, err := createAndOpenFile(filename)
	if err != nil {
		panic(err)
	}

	return &logger{f: f}
}

func (t *logger) write(s string) {
	_, err := t.f.WriteString(s)
	if err != nil {
		panic(err)
	}
}

func createAndOpenFile(filename string) (*os.File, error) {
	// Create the directory if it doesn't exist
	dirname := filepath.Dir(filename)
	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		if err = os.MkdirAll(dirname, 0755); err != nil {
			return nil, err
		}
	}

	// Truncate the file if it already exists
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, err
	}

	// Seek to the end of the file so writes will be appended
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}

	return f, nil
}
