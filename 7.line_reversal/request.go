package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Msg struct {
	mType  string
	sessID int

	// data contains the actual payload sent in the data packet
	data string

	// pos is a non-negative integer representing the position in the stream that the DATA belongs
	//
	// It refers to the position in the stream of unescaped application-layer bytes,
	// not the escaped data passed in LRCP.
	pos int32

	// ackLen is a non-negative integer telling the other side
	// how many bytes of payload have been successfully received so far
	ackLen int32
}

// Packet contents must begin with a forward slash, end with a forward slash,
// have a valid message type,
// and have the correct number of fields for the message type.
//
// Numeric field values must be smaller than 2147483648. This means sessions are limited to 2 billion bytes of data transferred in each direction.
// LRCP messages must be smaller than 1000 bytes. You might have to break up data into multiple data messages in order to fit it below this limit.
func parseMsg(msg string) (*Msg, error) {
	if len(msg) == 0 || len(msg) > 1000 {
		return nil, errors.New("length invalid")
	}

	if msg[0] != '/' || msg[len(msg)-1] != '/' {
		return nil, errors.New("not wraped properly with forward slash")
	}

	// Remove first and last forward slash
	msg = msg[1 : len(msg)-1]
	splits := strings.Split(unescapeData(msg), "/")

	if len(splits) < 2 {
		return nil, fmt.Errorf("invalid msg. insufficient sections. [%s]", msg)
	}

	sessID, err := strconv.Atoi(splits[1])
	if err != nil {
		return nil, fmt.Errorf("session id must be a number. [%s]", splits[1])
	}

	m := Msg{
		mType:  splits[0],
		sessID: sessID,
	}

	switch m.mType {
	case "connect", "close":
		if len(splits) != 2 {
			return nil, fmt.Errorf("connect:: invalid connection msg [%s]", msg)
		}

	case "data":
		if len(splits) < 4 {
			return nil, fmt.Errorf("data:: invalid connection msg [%s]", msg)
		}

		pos, err := strconv.Atoi(splits[2])
		if err != nil {
			return nil, fmt.Errorf("data:: position must be a number. [%s]", splits[1])
		}

		if pos < 0 {
			return nil, fmt.Errorf("data:: position must be positive. [%d]", pos)
		}
		m.pos = int32(pos)

		m.data = strings.Join(splits[3:], "/")

	case "ack":
		if len(splits) != 3 {
			return nil, fmt.Errorf("ack:: invalid connection msg [%s]", msg)
		}

		ackLen, err := strconv.Atoi(splits[2])
		if err != nil {
			return nil, fmt.Errorf("ack:: length must be a number. [%s]", splits[1])
		}
		m.ackLen = int32(ackLen)

	default:
		return nil, fmt.Errorf("invalid msg type [%s]", m.mType)
	}

	return &m, nil
}

func escapeData(m string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(m, `\`, `\\`),
		`/`, `\/`,
	)
}

func unescapeData(m string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(m, `\/`, `/`),
		`\\`, `\`,
	)
}
