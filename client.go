package torbit

type client interface {
	getName() string
	getRoom() string
	setRoom(room string)
	read()
	write(msg string) error
	close()
}

type _client struct {
	name string
	rwc  readWriteCloser
}

type readWriteCloser interface {
	read()
	write(msg string)
	close()
}
