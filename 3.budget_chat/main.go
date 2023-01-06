package main

import (
	"bufio"
	"log"
	"net"
	"strings"
)

func main() {
	addr := ":3723"
	network := "tcp"
	laddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		log.Fatalf("net.ResolveTCPAddr(); %v", err)
	}

	listener, err := net.ListenTCP(network, laddr)
	if err != nil {
		log.Fatalf("net.ListenTCP(); %v", err)
	}
	defer listener.Close()

	chatroom := newChatroom()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("listener.AcceptTCP(); %v", err)
			continue
		}

		go handleConn(conn, chatroom)
	}
}

const (
	maxUsernameLen = 20
)

func handleConn(conn *net.TCPConn, chatroom *Chatroom) {
	defer conn.Close()

	// 1. Set the username
	conn.Write([]byte("Welcome to budgetchat! What shall I call you?\n"))
	usernameB := make([]byte, maxUsernameLen)
	read, err := conn.Read(usernameB)
	if err != nil {
		return
	}

	username := strings.TrimSpace(string(usernameB[:read]))
	if !chatroom.validateUsername(username) {
		conn.Write([]byte("bad username: must contain at least 1 character, and must consist entirely of alphanumeric characters\n"))
		return
	}

	if err := chatroom.addUser(username, conn); err != nil {
		conn.Write([]byte("Username already taken :(\n"))
		return
	}
	defer func() {
		chatroom.announceDeparture(username)
		chatroom.removeUser(username)
	}()

	// 2. Announce participants
	chatroom.sendParticipants(username)

	// 3. Announce user join
	chatroom.announceUserJoin(username)

	// 4. Handle chat
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		msg := scanner.Text()
		chatroom.sendMsg(username, msg)
	}
}
