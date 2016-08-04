package torbit

import (
	"bufio"
	"net"
	"time"
)

const chatHelp = `(chatbot): Hello, welcome to the chat room
Commands:
  /help    see this help message again (example: /help)
  /rooms   list all rooms              (example: /rooms)
  /join    join a room                 (example: /join general)
  /newroom create a new room           (example: /newroom random)

`

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
			return err
		}
		// handle commands here to determine:
		//   1. text (duh)
		//   2. messageType
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
