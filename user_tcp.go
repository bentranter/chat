package torbit

import (
	"bufio"
	"net"
	"strings"
)

const chatHelp = `Hello, welcome to the chat server!
Commands:
  /help       see this help message again    (example: /help)
  /listusers  see all users connected        (example: /listusers)
  /listrooms  see all channels               (example: /listrooms)
  /newroom    create a new room and join it  (example: /newroom random)
  /join       join a room                    (example: /join random)
  /leave      leave a room                   (example: /leave random)
  /mute       mute a user                    (example: /mute rob)
  /unmute     unmute a user                  (example: /unmute rob)
  /mutes      see who you've muted           (example: /mutes)
  /dm         send a message to a user       (example: /dm rob: hello!)
`

type command func(tc *tcpUser, arg string)

// commands are each action a client is able to perform besides just sending
// plaintext.
var commands = map[string]command{
	"/help":      helpCmd,
	"/listusers": listUsersCmd,
	"/listrooms": listRoomsCmd,
	"/newroom":   newRoomCmd,
	"/join":      joinRoomCmd,
	"/leave":     leaveRoomCmd,
	"/mute":      muteCmd,
	"/unmute":    unmuteCmd,
	"/mutes":     mutesCmd,
	"/dm":        dmCmd,
}

// a tcpUser represents a telnet user, relying on text-only commands to
// communicate.
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
	r := bufio.NewReader(conn)
	conn.Write([]byte("Please enter your username: "))

	for {
		n, err := r.ReadString('\n')
		if err != nil {
			conn.Close()
		}
		n = strings.TrimSpace(n)
		if n == "" {
			conn.Write([]byte("Your name cannot be blank. Try again: "))
			continue
		}
		if _, ok := h.users[n]; !ok {
			name = n
			break
		}
		conn.Write([]byte("Sorry, the name " + n + " is already taken. Please choose another one: "))
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
	if _, ok := tc.muted[message.Username]; ok {
		return nil
	}
	switch message.MessageType {
	case text:
		return tc.writeText("(" + message.Username + " to " + message.Channel + "): " + message.Text)

	case listUsers, listChannels:
		return tc.writeText(message.Text + "\n")

	case join, create:
		tc.currentRoomName = message.Channel
		return tc.writeText(message.Text)

	case leave:
		tc.currentRoomName = defaultChannelName
		return tc.writeText(message.Text)

	case mute:
		if _, ok := tc.muted[message.Channel]; !ok {
			tc.muted[message.Channel] = true
			return tc.writeText(message.Text)
		}
		return tc.writeText("User " + message.Channel + " is already muted.\n")

	case unmute:
		if _, ok := tc.muted[message.Channel]; ok {
			delete(tc.muted, message.Channel)
			return tc.writeText(message.Text)
		}
		return tc.writeText("User " + message.Channel + " isn't muted.\n")

	case dm:
		tc.writeText("(" + message.Username + " to " + message.Channel + "): " + message.Text + "\n")
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

func dmCmd(tc *tcpUser, arg string) {
	i := strings.Index(arg, ":")
	if i == -1 {
		tc.writeText("/dm command not understood, you're missing a ':' from your command.\n")
		return
	}
	msgs := strings.SplitAfterN(arg, ":", 2)
	// sanity check, SplitAfterN should always return a length 2 slice
	if len(msgs) < 2 {
		tc.writeText("/dm command not understood, commands appears to be malformed. Type '/help' to see how to use each command.\n")
		return
	}

	user := strings.TrimSpace(strings.TrimRight(msgs[0], ":"))
	if user == "" {
		tc.writeText("/dm command not understood, you're missing a username.\n")
		return
	}

	if user == tc.username {
		tc.writeText("You can't send a dm to yourself.\n")
		return
	}

	msg := strings.TrimSpace(msgs[1])
	if msg == "" {
		tc.writeText("/dm command not understood, it looks like your message is blank.\n")
		return
	}
	tc.send <- (newMessage(user, tc.username, msg, dm))
}

func listUsersCmd(tc *tcpUser, arg string) {
	if arg != "" {
		tc.send <- newMessage(arg, tc.username, "", listUsers)
		return
	}
	tc.send <- newMessage("", tc.username, "", listUsers)
}

func listRoomsCmd(tc *tcpUser, _ string) {
	tc.send <- newMessage("", tc.username, "", listChannels)
}
