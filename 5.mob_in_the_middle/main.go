package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
)

const (
	addr           = ":3723"
	network        = "tcp"
	tonyAddress    = "7YWHMfk9JZe0LM0g1ZauHuiSxhI"
	upstreamServer = "chat.protohackers.com:16963"
)

func main() {
	listener, err := net.Listen(network, addr)
	if err != nil {
		log.Fatalf("net.Listen(); %v", err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("listener.Accept(); %v", err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	// Establish connection with the upstream server
	upstream, err := net.Dial(network, upstreamServer)
	if err != nil {
		log.Fatalf("net.Dial(); %v", err)
	}
	defer upstream.Close()

	go proxy(upstream, conn)
	proxy(conn, upstream)
}

func proxy(from, to net.Conn) {
	// The reason I'm not using bufio.NewScanner
	// is because I don't want any message that's not delimited
	// by "\n"
	s := bufio.NewReader(from)
	for {
		txt, err := s.ReadString('\n')
		if err != nil {
			log.Println(err)
			return
		}

		txt = strings.TrimSpace(txt)
		responseMsg, replaced := replaceAddress(txt)
		if replaced {
			log.Println("Attack success !!")
		}

		responseMsg += "\n"
		if _, err := to.Write([]byte(responseMsg)); err != nil {
			fmt.Println("Err writing", err)
			return
		}
	}
}

// A substring is considered to be a Boguscoin address if it satisfies all of:
// it starts with a "7"
// it consists of at least 26, and at most 35, alphanumeric characters
// it starts at the start of a chat message, or is preceded by a space
// it ends at the end of a chat message, or is followed by a space
func isBogusCoinAddress(s string) bool {
	return len(s) >= 26 && len(s) <= 35 && s[0] == '7'
}

func replaceAddress(msg string) (string, bool) {
	if len(msg) == 0 {
		return msg, false
	}

	var found bool
	splits := strings.Split(msg, " ")
	for i, s := range splits {
		if isBogusCoinAddress(s) {
			splits[i] = tonyAddress
			found = true
		}
	}

	if found {
		return strings.Join(splits, " "), true
	}

	return msg, false
}
