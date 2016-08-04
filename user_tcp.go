package torbit

import (
	"bufio"
	"net"
	"strings"
	"time"
)

const chatHelp = `(chatbot): Hello, welcome to the chat room
Commands:
  /help    see this help message again (example: /help)
  /rooms   list all rooms              (example: /rooms)
  /join    join a room                 (example: /join general)
  /newroom create a new room           (example: /newroom random)

`

type command func(tc *tcpUser, arg string)

var commands = map[string]command{
	"/help": helpCmd,
}

type tcpUser struct {
	currentRoomName string
	username        string
	r               *bufio.Reader
	w               *bufio.Writer
	conn            net.Conn
	receiver        chan *message
}

func newTCPUser(conn net.Conn, receiver chan *message) *tcpUser {
	// in here, read name, etc
	return &tcpUser{
		currentRoomName: defaultChannelName,
		username:        "pee",
		r:               bufio.NewReader(conn),
		w:               bufio.NewWriter(conn),
		conn:            conn,
		receiver:        receiver,
	}
}

func (tc *tcpUser) read() error {
	for {
		messageText, err := tc.r.ReadString('\n')
		if err != nil {
			println(err.Error())
			return err
		}
		if ok := tc.handleCommand(messageText); ok {
			continue
		}
		tc.receiver <- &message{
			channel:     tc.currentRoomName,
			username:    tc.username,
			text:        messageText,
			time:        time.Now(),
			messageType: text,
		}
	}
}

func (tc *tcpUser) write(message string) error {
	_, err := tc.w.WriteString(message)
	if err != nil {
		return err
	}
	return tc.w.Flush()
}

func (tc *tcpUser) close() {
	tc.conn.Close()
}

func (tc *tcpUser) name() string {
	return tc.username
}

func (tc *tcpUser) setCurrentRoom(roomName string) {
	tc.currentRoomName = roomName
}

func (tc *tcpUser) handleCommand(s string) bool {
	if !strings.HasPrefix(s, "/") {
		return false
	}
	cmd := strings.TrimSpace(strings.Split(s, " ")[0])
	cmdFunc, ok := commands[cmd]
	if !ok {
		tc.write("Command " + cmd + " doesn't exist.\n")
		return true
	}
	cmdArg := strings.TrimSpace(strings.TrimPrefix(s, cmd))
	cmdFunc(tc, cmdArg)
	return true
}

func helpCmd(tc *tcpUser, _ string) {
	tc.write(chatHelp)
}
