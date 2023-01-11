package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	network := "udp"
	addr := ":3723"

	conn, err := net.ListenPacket(network, addr)
	if err != nil {
		log.Fatalf("net.ListenPacket(); %v", err)
	}
	defer conn.Close()

	log.Printf("Listening to udp conns in [%v]\n", addr)

	var b = make([]byte, 1000)
	var store = newStore()
	for {
		read, addr, err := conn.ReadFrom(b)
		if err != nil {
			log.Printf("Err:: conn.ReadFromUDP(); %v", err)
			continue
		}

		handlePacket(store, conn, addr, string(b[:read]))
	}
}

func handlePacket(store *store, conn net.PacketConn, remoteAddr net.Addr, m string) {
	msg, err := parseMsg(m)
	if err != nil {
		log.Println("validateMsg", err)
		return
	}

	log.Println("<--", m)
	s := store.Session(msg.sessID)

	switch msg.mType {
	case "connect":
		s = store.OpenSession(msg.sessID, remoteAddr)
		sendAck(s, conn, remoteAddr, msg.sessID, 0)

	case "data":
		if s == nil {
			closeSess(store, conn, remoteAddr, msg.sessID)
			return
		}

		sendAck(s, conn, remoteAddr, msg.sessID, len(msg.data))

		// Send data
		if msg.data[len(msg.data)-1] == '\n' {
			sendData(s, conn, remoteAddr, msg.sessID, len(msg.data), reverse(msg.data[0:len(msg.data)-1]))
		}

	case "ack":
		if s == nil {
			closeSess(store, conn, remoteAddr, msg.sessID)
			return
		}

		if msg.ackLen < s.maxAck {
			// do nothing
			// assume it's a duplicate ack that got delayed
		}

		if msg.ackLen > s.bSent {
			// peer is misbehaving, close the session
			closeSess(store, conn, remoteAddr, msg.sessID)
			return
		}

		// If the LENGTH value is smaller than the total amount of payload you've sent:
		// retransmit all payload data after the first LENGTH bytes
		if msg.ackLen < s.bSent {
			sendData(s, conn, remoteAddr, s.id, 0, "olleh")
		}

		// If the LENGTH value is equal to the total amount of payload you've sent: don't send any reply
		if msg.ackLen == s.bSent {
		}

	case "close":
		closeSess(store, conn, remoteAddr, msg.sessID)
		return
	}
}

func closeSess(store *store, c net.PacketConn, addr net.Addr, sessID int) {
	_, err := c.WriteTo([]byte(fmt.Sprintf("/close/%d/", sessID)), addr)
	if err != nil {
		fmt.Println(err)
	}

	store.Close(sessID)
}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func sendAck(s *session, c net.PacketConn, addr net.Addr, sessID, length int) {
	s.Ack(sessID, int32(length))
	ack := fmt.Sprintf("/ack/%d/%d/", sessID, length)
	log.Println("-->", ack)

	_, err := c.WriteTo([]byte(ack), addr)
	if err != nil {
		fmt.Println(err)
	}
}

func sendData(s *session, c net.PacketConn, addr net.Addr, sessID, pos int, data string) {
	s.RegisterPayload(pos, data)
	msg := fmt.Sprintf("/data/%d/%d/%s\n/", sessID, pos, data)
	log.Println("-->", msg)

	_, err := c.WriteTo([]byte(msg), addr)
	if err != nil {
		fmt.Println(err)
	}
}
