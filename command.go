package torbit

import (
	"strings"
)

func handleCommand(c client, msg string) bool {
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
	cmdFunc(c, cmdArg)
	return true
}

type command func(c client, arg string)

var commands = map[string]command{
	"/help": helpCmd,
	"/join": joinRoomCmd,
}

func helpCmd(c client, _ string) {
	c.write(chatHelp)
}

func joinRoomCmd(c client, arg string) {
	if arg == "" {
		c.write("Room name cannot be empty\n")
		return
	}
	ch := c.roomChangeCh()
	ch <- &roomChange{
		newRoomName: arg,
		c:           c,
	}
}
