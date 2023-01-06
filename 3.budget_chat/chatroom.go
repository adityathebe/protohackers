package main

import (
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"sync"
)

type Chatroom struct {
	m              *sync.Mutex
	Users          map[string]net.Conn
	usernameRegexp *regexp.Regexp
}

func newChatroom() *Chatroom {
	return &Chatroom{
		m:              &sync.Mutex{},
		Users:          make(map[string]net.Conn),
		usernameRegexp: regexp.MustCompile(`^[a-zA-Z0-9]+$`),
	}
}

func (t *Chatroom) addUser(u string, conn net.Conn) error {
	t.m.Lock()
	if _, ok := t.Users[u]; ok {
		return errors.New("username taken")
	}

	t.Users[u] = conn
	t.m.Unlock()
	return nil
}

func (t *Chatroom) validateUsername(username string) bool {
	return len(username) > 0 && t.usernameRegexp.MatchString(username)
}

func (t *Chatroom) removeUser(u string) {
	t.m.Lock()
	delete(t.Users, u)
	t.m.Unlock()
}

func (t *Chatroom) getParticipants(exceptions ...string) []string {
	users := make([]string, 0, len(t.Users))
	for activeUser := range t.Users {
		if !isInSlice(exceptions, activeUser) {
			users = append(users, activeUser)
		}
	}

	return users
}

func (t *Chatroom) announceUserJoin(joinee string) {
	welcomeMsg := fmt.Sprintf("* %s has entered the room\n", joinee)
	for activeUser, conn := range t.Users {
		if activeUser != joinee {
			conn.Write([]byte(welcomeMsg))
		}
	}
}

func (t *Chatroom) sendMsg(username, msg string) {
	welcomeMsg := fmt.Sprintf("[%s] %s\n", username, msg)
	for activeUser, conn := range t.Users {
		if activeUser != username {
			conn.Write([]byte(welcomeMsg))
		}
	}
}

func (t *Chatroom) announceDeparture(username string) {
	welcomeMsg := fmt.Sprintf("* %s has left the room\n", username)
	for activeUser, conn := range t.Users {
		if activeUser != username {
			conn.Write([]byte(welcomeMsg))
		}
	}
}

func (t *Chatroom) sendParticipants(joinee string) {
	participants := t.getParticipants(joinee)
	presenceNotice := fmt.Sprintf("* The room contains: %s\n", strings.Join(participants, ", "))
	t.Users[joinee].Write([]byte(presenceNotice))
}

func isInSlice(s []string, item string) bool {
	for _, sItem := range s {
		if item == sItem {
			return true
		}
	}

	return false
}
