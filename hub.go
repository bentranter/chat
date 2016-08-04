package torbit

import (
	"fmt"
	"log"
	"net"
	"time"
)

type messageType int

const (
	defaultChannelName = "general"

	join = messageType(iota)
	create
	leave
	text
	quit
)

type message struct {
	channel     string
	username    string
	text        string
	time        time.Time
	messageType messageType
}

func newMessage(channel, username, text string, messageType messageType) *message {
	return &message{
		channel:     channel,
		username:    username,
		text:        text,
		time:        time.Now(),
		messageType: messageType,
	}
}

type channel struct {
	name        string
	users       map[*User]bool
	activeUsers map[string]*User
	broadcastCh chan *message
}

func newChannel(channelName string, activeUsers map[string]*User) *channel {
	ch := &channel{
		name:        channelName,
		users:       make(map[*User]bool),
		activeUsers: make(map[string]*User),
		broadcastCh: make(chan *message),
	}
	ch.activeUsers = activeUsers
	go ch.broadcast()
	return ch
}

func (c *channel) join(u *User) {
	c.users[u] = true
	c.broadcastCh <- newMessage(c.name, u.name, u.name+" has joined\n", text)
}

func (c *channel) leave(u *User) {
	delete(c.users, u)
}

func (c *channel) broadcast() {
	for {
		msg := <-c.broadcastCh
		for u := range c.users {
			if _, ok := c.activeUsers[u.name]; !ok {
				continue
			}
			err := u.conn.write(fmt.Sprintf("(%s to %s): %s", msg.username, msg.channel, msg.text))
			if err != nil {
				println("Broadcast error: ", err.Error())
			}
		}
	}
}

type hub struct {
	logger    *log.Logger
	channels  map[string]*channel
	users     map[string]*User
	userCh    chan *User
	messageCh chan *message
}

func newHub(l *log.Logger) *hub {
	return &hub{
		logger:    l,
		channels:  make(map[string]*channel),
		users:     make(map[string]*User),
		userCh:    make(chan *User),
		messageCh: make(chan *message),
	}
}

func (h *hub) run() {
	h.channels[defaultChannelName] = newChannel(defaultChannelName, h.users)
	for {
		select {
		case user := <-h.userCh:
			h.users[user.name] = user
			h.channels[defaultChannelName].join(user)
			go user.conn.read()

		case message := <-h.messageCh:
			switch message.messageType {

			case join:
				// need to check for nil entries
				h.channels[message.channel].users[h.users[message.username]] = true

			case create:
				h.channels[message.channel] = newChannel(message.channel, h.users)
				h.channels[message.channel].users[h.users[message.username]] = true

			case leave:
				delete(h.channels[message.channel].users, h.users[message.username])

			case text:
				h.logger.Printf("(%s to %s): %s", message.username, message.channel, message.text)
				h.channels[message.channel].broadcastCh <- message

			case quit:
				h.logger.Printf("%s", message.text)
				h.users[message.username].conn.close() // pls panic
				delete(h.users, message.username)
				h.channels[defaultChannelName].broadcastCh <- message
			}
		}
	}
}

func (h *hub) serve(port string) error {
	server, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	h.logger.Println("Server started on", port)

	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				h.logger.Println(err.Error())
			}
			newUser := createTCPUser(conn, h)
			h.userCh <- newUser
		}
	}()

	h.run()
	return nil
}

func ListenAndServe(l *log.Logger, port string) error {
	h := newHub(l)
	return h.serve(port)
}
