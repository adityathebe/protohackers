package main

import (
	"encoding/binary"
	"io"
)

type RequestType uint8

const (
	Plate         RequestType = 0x20
	WantHeartbeat RequestType = 0x40
	IAmCamera     RequestType = 0x80
	IAmDispatcher RequestType = 0x81
)

type CameraReq struct {
	road  uint16 // road field contains the road number that the camera is on
	mile  uint16 // mile contains the position of the camera
	limit uint16 // limit contains the speed limit of the road
}

type DispatcherReq struct {
	numroads uint8
	roads    []uint16
}

type PlateData struct {
	name      string
	timestamp uint32
}

type Request struct {
	Type      RequestType
	Camera    CameraReq
	Plate     PlateData
	Dispatch  DispatcherReq
	Heartbeat uint32
}

func DecodeRequest(reader io.Reader) (Request, error) {
	var r Request
	if err := binary.Read(reader, binary.BigEndian, &r.Type); err != nil {
		return r, err
	}

	switch r.Type {
	case Plate:
		var size byte
		if err := binary.Read(reader, binary.BigEndian, &size); err != nil {
			return r, err
		}

		var plateName = make([]uint8, size)
		if err := binary.Read(reader, binary.BigEndian, &plateName); err != nil {
			return r, err
		}
		r.Plate.name = string(plateName)

		if err := binary.Read(reader, binary.BigEndian, &r.Plate.timestamp); err != nil {
			return r, err
		}

	case WantHeartbeat:
		if err := binary.Read(reader, binary.BigEndian, &r.Heartbeat); err != nil {
			return r, err
		}

	case IAmCamera:
		if err := binary.Read(reader, binary.BigEndian, &r.Camera.road); err != nil {
			return r, err
		}
		if err := binary.Read(reader, binary.BigEndian, &r.Camera.mile); err != nil {
			return r, err
		}
		if err := binary.Read(reader, binary.BigEndian, &r.Camera.limit); err != nil {
			return r, err
		}

	case IAmDispatcher:
		if err := binary.Read(reader, binary.BigEndian, &r.Dispatch.numroads); err != nil {
			return r, err
		}

		r.Dispatch.roads = make([]uint16, r.Dispatch.numroads)
		if err := binary.Read(reader, binary.BigEndian, &r.Dispatch.roads); err != nil {
			return r, err
		}
	}

	return r, nil
}
