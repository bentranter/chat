package torbit

import (
	"strings"
)

func handleCommand(s *server, c client, msg string) bool {
	if !strings.HasPrefix(msg, "/") {
		return false
	}

	cmd := strings.TrimSpace(strings.Split(msg, " ")[0]) // get first arg
	cmdFunc, ok := commands[cmd]
	if !ok {
		c.write("(chatbot): " + strings.TrimSpace(msg) + " isn't a command. Type /help to see available commands\n")
		return true // command doesn't exist, but it's valid command syntax
	}

	cmdArg := strings.TrimSpace(strings.TrimPrefix(msg, cmd))
	cmdFunc(s, c, cmdArg)
	return true
}

type command func(s *server, c client, arg string)

var commands = map[string]command{
	"/help":    helpCmd,
	"/join":    joinRoomCmd,
	"/newroom": newRoomCmd,
}

func helpCmd(_ *server, c client, _ string) {
	c.write(chatHelp)
}

func joinRoomCmd(s *server, c client, arg string) {
	if arg == "" {
		c.write("Room name cannot be empty\n")
		return
	}
	err := s.changeRoom(&roomChange{
		newRoomName: arg,
		c:           c,
	})
	if err != nil {
		c.write(err.Error())
		return
	}
	c.write("Joined room " + arg + ".\n")
}

func newRoomCmd(s *server, c client, arg string) {
	if arg == "" {
		c.write("New room name cannot be empty\n")
		return
	}
	err := s.newRoom(arg)
	if err != nil {
		c.write(err.Error())
		return
	}
	c.write(("(chatbot): New room " + arg + " created successfully\n"))
}
