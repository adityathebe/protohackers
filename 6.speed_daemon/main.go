package main

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/adityathebe/protohackers"
)

var (
	store           *Store
	dispatcherStore *DispatcherStore
)

// Cameras report to you
// Detect any speed limit violation &
// report to the responsible ticket dispatchers
func main() {
	store = newStore()
	dispatcherStore = newDispatcherStore()
	protohackers.StartTCPServer(handleConn)
}

type clientIdentity struct {
	dispatcherID int
	isCamera     bool
	Camera       CameraReq
	Dispatch     DispatcherReq
}

func handleConn(conn *net.TCPConn) {
	var (
		identity           *clientIdentity
		heartbeatRequested bool
	)

	defer func() {
		if identity != nil && !identity.isCamera {
			dispatcherStore.Unregister(identity.dispatcherID)
		}
	}()

	for {
		r, err := DecodeRequest(conn)
		if err != nil {
			log.Printf("DecodeRequest(); [%v] %v\n", r, err)
			return
		}

		// Identify first
		if identity == nil && r.Type != WantHeartbeat {
			switch r.Type {
			case IAmCamera:
				identity = &clientIdentity{isCamera: true, Camera: r.Camera}
				log.Printf("Identified as Camera 📷. [road=%d mile=%d limit=%d]\n", r.Camera.road, r.Camera.mile, r.Camera.limit)
			case IAmDispatcher:
				identity = &clientIdentity{dispatcherID: dispatcherStore.getDispatcherID(), isCamera: false, Dispatch: r.Dispatch}
				log.Printf("Identified as Dispatcher 🎫. [%v]\n", r.Dispatch)
			default:
				sendErr(conn, "please identify yourself")
				return
			}
		}

		switch r.Type {
		case Plate:
			if !identity.isCamera {
				sendErr(conn, "only cameras can send plates")
				return
			}

			store.SavePlate(identity.Camera.road, identity.Camera.mile, r.Plate.timestamp, r.Plate.name)

			if ticketMsg := generateTicket(identity.Camera.road, r.Plate.name); ticketMsg != nil {
				dispatcherStore.Dispatch(ticketMsg)
			}

		case WantHeartbeat:
			if heartbeatRequested {
				sendErr(conn, "It is an error for a client to send multiple WantHeartbeat messages on a single connection")
				return
			}

			if r.Heartbeat != 0 {
				go func() {
					res := Response{Type: Heartbeat}
					duration := getDurationFromDeciseconds(r.Heartbeat)
					for {
						if _, err := conn.Write(res.Encode()); err != nil {
							return
						}

						time.Sleep(duration)
					}
				}()
			}

		case IAmCamera:
			if !identity.isCamera {
				sendErr(conn, "you are a ticket dispatcher")
				return
			}

			store.SaveSpeedLimit(r.Camera.road, r.Camera.limit)

		case IAmDispatcher:
			if identity.isCamera {
				sendErr(conn, "you are a camera")
				return
			}

			dispatcherStore.Register(identity.dispatcherID, conn, r.Dispatch.roads)
		}
	}
}

func sendErr(c io.Writer, msg string) {
	log.Println("Err", msg)
	m := Response{Type: Error, ErrMsg: msg}
	c.Write(m.Encode())
}

func ParseRequest(stream []byte, idx int) (*Request, error) {
	return nil, nil
}

func generateTicket(road uint16, plate string) []TicketMsg {
	platesOfRoad := store.PlatesOfRoad(road)
	speedLimit := store.SpeedLimit(road)
	plateRecords := platesOfRoad[plate]
	ticketMsg := detectOverSpeeding(road, speedLimit, plate, plateRecords)
	return ticketMsg
}

// detectOverSpeeding checks for overspeeding of a particular plate at a particular road.
func detectOverSpeeding(road, speedLimit uint16, plateName string, plateRecords []PlateRecord) []TicketMsg {
	var overSpeed []TicketMsg
	for i := 1; i < len(plateRecords); i++ {
		distance := dist(plateRecords[i].mile, plateRecords[i-1].mile)
		duration := plateRecords[i].timestamp - plateRecords[i-1].timestamp
		mph := (float64(distance)) / (float64(duration) / 3600)
		if mph > float64(speedLimit) {
			overSpeed = append(overSpeed, TicketMsg{
				Plate:      plateName,
				Mile1:      plateRecords[i-1].mile,
				Timestamp1: plateRecords[i-1].timestamp,
				Mile2:      plateRecords[i].mile,
				Timestamp2: plateRecords[i].timestamp,
				Speed:      uint16(mph * 100),
				Road:       road,
			})
		}
	}

	return overSpeed
}

func dist(a, b uint16) uint16 {
	if a > b {
		return a - b
	}

	return b - a
}

func getDurationFromDeciseconds(ds uint32) time.Duration {
	return time.Duration(ds) * time.Millisecond * 100
}
