package main

import (
	"bufio"
	"net"
	"strings"

	"github.com/adityathebe/protohackers"
)

var chatroom = newChatroom()

func main() {
	protohackers.StartTCPServer(handleConn)
}

const (
	maxUsernameLen = 20
)

func handleConn(conn *net.TCPConn) {
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
