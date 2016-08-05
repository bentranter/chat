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
  /join    join a room                   (example: /join random)
  /leave   leave a room                  (example: /leave random)
  /mute    mute a user                   (example: /mute rob)
  /unmute  unmute a user                 (example: /unmute rob)
  /mutes   see who you've muted          (example: /mutes)
  @<user>  send a message to a user      (example: @rob hello!)
`

type command func(tc *tcpUser, arg string)

var commands = map[string]command{
	"/help":    helpCmd,
	"/newroom": newRoomCmd,
	"/join":    joinRoomCmd,
	"/leave":   leaveRoomCmd,
	"/mute":    muteCmd,
	"/unmute":  unmuteCmd,
	"/mutes":   mutesCmd,
}

type tcpUser struct {
	currentRoomName string
	muted           map[string]bool
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
		muted:           make(map[string]bool),
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
	if _, ok := tc.muted[message.username]; ok {
		return nil
	}
	switch message.messageType {
	case text:
		return tc.writeText("(" + message.username + " to " + message.channel + "): " + message.text)

	case join, create:
		tc.currentRoomName = message.channel
		return tc.writeText(message.text)

	case leave:
		tc.currentRoomName = defaultChannelName
		return tc.writeText(message.text)

	case mute:
		if _, ok := tc.muted[message.channel]; !ok {
			tc.muted[message.channel] = true
			return tc.writeText(message.text)
		}
		return tc.writeText("User " + message.channel + " is already muted.\n")

	case unmute:
		if _, ok := tc.muted[message.channel]; ok {
			delete(tc.muted, message.channel)
			return tc.writeText(message.text)
		}
		return tc.writeText("User " + message.channel + " isn't muted.\n")

	case dm:
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
	arg = strings.TrimSpace(arg)
	if arg == "" {
		tc.write(newMessage("you", "server", "Room name cannot be blank\n", text))
		return
	}
	if arg == tc.currentRoomName {
		tc.write(newMessage("you", "server", "You're already in that room\n", text))
		return
	}
	tc.send <- newMessage(arg, tc.username, tc.username+" created new channel "+arg, create)
}

func joinRoomCmd(tc *tcpUser, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		tc.write(newMessage("you", "server", "Room name cannot be blank\n", text))
		return
	}
	if arg == tc.currentRoomName {
		tc.write(newMessage("you", "server", "You're already in that room\n", text))
		return
	}
	tc.send <- newMessage(arg, tc.username, tc.username+" joined channel "+arg, join)
}

func leaveRoomCmd(tc *tcpUser, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		tc.write(newMessage("you", "server", "Room name cannot be blank\n", text))
	}
	tc.send <- newMessage(arg, tc.username, tc.username+" joined channel "+arg, leave)
}

func muteCmd(tc *tcpUser, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		tc.writeText("Username cannot be blank\n")
		return
	}
	if arg == tc.username {
		tc.writeText("You can't mute yourself\n")
		return
	}

	tc.send <- newMessage(arg, tc.username, "Muted user "+arg+".\n", mute)
}

func unmuteCmd(tc *tcpUser, arg string) {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		tc.writeText("Username cannot be blank\n")
		return
	}
	tc.send <- newMessage(arg, tc.username, "Unmuted user "+arg+".\n", unmute)
}

func mutesCmd(tc *tcpUser, _ string) {
	if len(tc.muted) < 1 {
		tc.writeText("You haven't muted anyone.\n")
		return
	}
	var mutes []string
	for mute := range tc.muted {
		mutes = append(mutes, mute)
	}
	muteList := strings.Join(mutes, "\n  - ")
	tc.writeText("You've muted:\n  - " + muteList + "\n")
}
