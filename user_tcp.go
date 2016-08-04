package torbit

import (
	"bufio"
	"net"
	"strings"
)

const chatHelp = `(chatbot to you): Hello, welcome to the chat room
Commands:
  /help    see this help message again (example: /help)
  /join    join a room                 (example: /join general)
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
		w:               bufio.NewWriter(conn),
		conn:            conn,
		receiver:        h.messageCh,
	}
}

func (tc *tcpUser) read() error {
	for {
		messageText, err := tc.r.ReadString('\n')
		if err != nil {
			tc.receiver <- newMessage("everyone", tc.username, tc.username+" has left that chat\n", quit)
			return err
		}
		if ok := tc.handleCommand(messageText); ok {
			continue
		}
		tc.receiver <- newMessage(tc.currentRoomName, tc.username, messageText, text)
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
