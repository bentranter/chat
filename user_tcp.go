package torbit

import (
	"bufio"
	"net"
	"strings"
)

const chatHelp = `(chatbot to you): Hello, welcome to the chat room
Commands:
  /help    see this help message again   (example: /help)
  /newroom create a new room and join it (example: /newroom random)
  /join    join a room                   (example: /join general)
`

type command func(tc *tcpUser, arg string)

var commands = map[string]command{
	"/help":    helpCmd,
	"/newroom": newRoomCmd,
	"/join":    joinRoomCmd,
}

type tcpUser struct {
	currentRoomName string
	username        string
	r               *bufio.Reader
	conn            net.Conn
	send            chan<- *message
}

func newTCPUser(conn net.Conn, h *hub) *tcpUser {
	var name string
	w := bufio.NewWriter(conn)
	r := bufio.NewReader(conn)
	w.WriteString("Please enter your username: ")
	w.Flush()

	for {
		n, err := r.ReadString('\n')
		if err != nil {
			conn.Close()
		}
		n = strings.TrimSpace(n)
		if _, ok := h.users[n]; !ok {
			name = n
			break
		}
		w.WriteString("(chatbot): Sorry, the name " + n + " is already taken. Please choose another one: ")
		w.Flush()
		continue
	}

	return &tcpUser{
		currentRoomName: defaultChannelName,
		username:        name,
		r:               bufio.NewReader(conn),
		conn:            conn,
		send:            h.messageCh,
	}
}

func (tc *tcpUser) read() error {
	for {
		messageText, err := tc.r.ReadString('\n')
		if err != nil {
			tc.send <- newMessage("everyone", tc.username, tc.username+" has left that chat\n", quit)
			return err
		}
		if ok := tc.handleCommand(messageText); ok {
			continue
		}
		tc.send <- newMessage(tc.currentRoomName, tc.username, messageText, text)
	}
}

func (tc *tcpUser) write(message *message) error {
	switch message.messageType {
	case text:
		return tc.writeText("(" + message.username + " to " + message.channel + "): " + message.text)

	case join, create:
		tc.currentRoomName = message.channel
		return tc.writeText(message.text)
	}
	return nil
}

func (tc *tcpUser) writeText(text string) error {
	_, err := tc.conn.Write([]byte(text))
	if err != nil {
		return err
	}
	return nil
}

func (tc *tcpUser) close() {
	tc.conn.Close()
}

func (tc *tcpUser) name() string {
	return tc.username
}

func (tc *tcpUser) handleCommand(s string) bool {
	if !strings.HasPrefix(s, "/") {
		return false
	}
	cmd := strings.TrimSpace(strings.Split(s, " ")[0])
	cmdFunc, ok := commands[cmd]
	if !ok {
		tc.write(newMessage(tc.currentRoomName, tc.username, "Command "+cmd+" doesn't exist\n", text))
		return true
	}
	cmdArg := strings.TrimSpace(strings.TrimPrefix(s, cmd))
	cmdFunc(tc, cmdArg)
	return true
}

func helpCmd(tc *tcpUser, _ string) {
	tc.write(newMessage("you", "server", chatHelp, text))
}

func newRoomCmd(tc *tcpUser, arg string) {
	if arg == tc.currentRoomName {
		tc.write(newMessage("you", "server", "You're already in that room\n", text))
		return
	}
	tc.send <- newMessage(arg, tc.username, tc.username+" created new channel "+arg, create)
}

func joinRoomCmd(tc *tcpUser, arg string) {
	if arg == tc.currentRoomName {
		tc.write(newMessage("you", "server", "You're already in that room\n", text))
		return
	}
	tc.send <- newMessage(arg, tc.username, tc.username+" joined channel "+arg, join)
}
