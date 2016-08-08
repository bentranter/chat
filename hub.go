package torbit

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

type messageType int

const (
	defaultChannelName = "general"

	join = messageType(iota)
	listUsers
	listChannels
	create
	leave
	text
	mute
	unmute
	dm
	quit
)

// A Config sets the options the server needs when it starts.
type Config struct {
	TCPPortAddr   string
	TCPSPortAddr  string
	HTTPPortAddr  string
	HTTPSPortAddr string
	IPAddr        string
	LogFilename   string
}

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
}

func newChannel(channelName string, activeUsers map[string]*User) *channel {
	return &channel{
		name:        channelName,
		users:       make(map[*User]bool),
		activeUsers: activeUsers,
	}
}

func (c *channel) join(u *User) {
	if _, ok := c.users[u]; ok {
		u.conn.write(newMessage(c.name, u.name, "Changing to channel "+c.name+"\n", join))
		return
	}
	c.users[u] = true
	c.broadcast(newMessage(c.name, u.name, u.name+" has joined "+c.name+"\n", join))
}

func (c *channel) leave(u *User) {
	delete(c.users, u)
}

func (c *channel) broadcast(m *message) {
	for u := range c.users {
		if _, ok := c.activeUsers[u.name]; !ok {
			continue
		}
		err := u.conn.write(m)
		if err != nil {
			// handle this didderefent
			println("Broadcast error: ", err.Error())
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

func (h *hub) listUsers(m *message) {
	user, ok := h.users[m.username]
	if !ok {
		return
	}
	var users []string
	if m.channel != "" {
		ch, ok := h.channels[m.channel]
		if !ok {
			user.conn.write(newMessage("you", "server", "Channel "+m.channel+" doesn't exist.\n", text))
			return
		}
		for u := range ch.users {
			users = append(users, u.name)
		}
	} else {
		for user := range h.users {
			users = append(users, user)
		}
	}
	m.text = strings.Join(users, ",")
	user.conn.write(m)
}

func (h *hub) listChannels(m *message) {
	user, ok := h.users[m.username]
	if !ok {
		return
	}
	var chans []string
	for ch := range h.channels {
		chans = append(chans, ch)
	}
	m.text = strings.Join(chans, ",")
	user.conn.write(m)
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
	m.text = "Sorry, the channel " + m.channel + " doesn't exist.\n"
	m.messageType = text
	user.conn.write(m)
}

func (h *hub) leaveChannel(m *message) {
	user, ok := h.users[m.username]
	if !ok {
		return
	}
	if m.channel == defaultChannelName {
		m.text = "You can't leave the default channel (which is " + defaultChannelName + ").\n"
		m.messageType = text
		user.conn.write(m)
		return
	}
	ch, ok := h.channels[m.channel]
	if !ok {
		m.text = "The channel " + m.channel + " doesn't exist, so you can't leave it.\n"
		m.messageType = text
		user.conn.write(m)
		return
	}
	if _, ok = ch.users[user]; !ok {
		m.text = "You're not a member of the channel " + m.channel + ".\n"
		m.messageType = text
		user.conn.write(m)
		return
	}
	ch.leave(user)
	m.text = "Left channel " + m.channel + ". Returning you to the general channel.\n"
	user.conn.write(m)
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

func (h *hub) broadcast(m *message) {
	h.logger.Printf("(%s to %s): %s", m.username, m.channel, m.text)
	ch, ok := h.channels[m.channel]
	if !ok {
		return
	}
	ch.broadcast(m)
}

func (h *hub) mute(m *message) {
	user, ok := h.users[m.username]
	if !ok {
		return
	}
	if _, ok := h.users[m.channel]; !ok {
		m.messageType = text
		m.text = "The user " + m.channel + " doesn't exist.\n"
		user.conn.write(m)
		return
	}
	user.conn.write(m)
}

func (h *hub) unmute(m *message) {
	user, ok := h.users[m.username]
	if !ok {
		return
	}
	if _, ok := h.users[m.channel]; !ok {
		m.messageType = text
		m.text = "The user " + m.channel + " doesn't exist.\n"
		user.conn.write(m)
		return
	}
	user.conn.write(m)
}

func (h *hub) dm(m *message) {
	h.logger.Printf("(%s to %s): %s", m.username, m.channel, m.text)
	sender, ok := h.users[m.username]
	if !ok {
		return
	}
	recipient, ok := h.users[m.channel]
	if !ok {
		m.messageType = text
		m.text = "Sorry, the user " + m.channel + " doesn't exist.\n"
		sender.conn.write(m)
		return
	}
	recipient.conn.write(m)
	sender.conn.write(m)
}

func (h *hub) quit(m *message) {
	h.logger.Printf("(%s to %s): %s", m.username, m.channel, m.text)
	user, ok := h.users[m.username]
	if !ok {
		return
	}
	user.conn.close()
	delete(h.users, m.username)
	h.channels[defaultChannelName].broadcast(m)
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
			switch message.messageType {

			case join:
				h.joinChannel(message)

			case listUsers:
				h.listUsers(message)

			case listChannels:
				h.listChannels(message)

			case create:
				h.createChannel(message)

			case leave:
				h.leaveChannel(message)

			case text:
				h.broadcast(message)

			case mute:
				h.mute(message)

			case unmute:
				h.unmute(message)

			case dm:
				h.dm(message)

			case quit:
				h.quit(message)
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
			go func() {
				h.userCh <- createTCPUser(conn, h)
			}()
		}
	}()

	h.run()
	return nil
}

func (h *hub) serveSecure(port string) error {
	server, err := tls.Listen("tcp", port, DefaultTLSConfig())
	if err != nil {
		h.logger.Println("Unable to start secure server:", err.Error())
		return err
	}
	h.logger.Println("Secure server started on", port)

	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				h.logger.Println(err.Error())
			}
			go func() {
				h.userCh <- createTCPUser(conn, h)
			}()
		}
	}()

	h.run()
	return nil
}

func (h *hub) serveHTTP(port string, mux http.Handler) error {
	h.logger.Println("HTTP server started on", port)
	return http.ListenAndServe(":"+port, mux)
}

func (h *hub) serveHTTPS(port string, mux http.Handler) error {
	server := &http.Server{
		Addr:      ":" + port,
		Handler:   mux,
		TLSConfig: DefaultTLSConfig(),
	}
	h.logger.Println("HTTPS server started on", port)
	return server.ListenAndServeTLS("", "")
}

func ListenAndServe(l *log.Logger, cfg *Config) error {
	h := newHub(l)
	errCh := make(chan error, 4)
	mux := getServeMux()

	go h.serveHTTP(cfg.HTTPPortAddr, mux)
	go h.serveHTTPS(cfg.HTTPSPortAddr, mux)
	go h.serve(":" + cfg.TCPPortAddr)
	go h.serveSecure(":" + cfg.TCPSPortAddr)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case err := <-errCh:
			if err != nil {
				h.logger.Fatalf("%s\n", err.Error())
			}
		case s := <-signalCh:
			log.Printf("Captured %v. Exiting...\n", s)
			os.Exit(1)
		}
	}
}
