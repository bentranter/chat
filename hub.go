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
	if _, ok := c.users[u]; ok {
		u.conn.write("Changing to channel " + c.name + "\n")
		return
	}
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

func (h *hub) newUser(u *User) {
	h.users[u.name] = u
	h.channels[defaultChannelName].join(u)
	go u.conn.read()
}

func (h *hub) joinChannel(m *message) {
	user, ok := h.users[m.username]
	if !ok {
		return
	}
	if ch, ok := h.channels[m.channel]; ok {
		ch.join(user)
		return
	}
	user.conn.write("Sorry, the channel " + m.channel + " doesn't exist.\n")
}

func (h *hub) createChannel(m *message) {
	user, ok := h.users[m.username]
	if !ok {
		return
	}
	ch, ok := h.channels[m.channel]
	if ok {
		ch.join(user)
		return
	}
	newCh := newChannel(m.channel, h.users)
	h.channels[m.channel] = newCh
	newCh.join(user)
}

func (h *hub) run() {
	h.channels[defaultChannelName] = newChannel(defaultChannelName, h.users)
	for {
		select {
		case user := <-h.userCh:
			// bug: if a user joins and then quits and re-joins with that same
			// name, you get a write to closed error
			h.newUser(user)

		case message := <-h.messageCh:
			h.logger.Printf("DEBUG: (%s to %s): %s", message.username, message.channel, message.text)
			switch message.messageType {

			case join:
				h.joinChannel(message)

			case create:
				// need to check for already created -- if so, then join it
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
