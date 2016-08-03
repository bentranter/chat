package torbit

type client interface {
	getName() string
	getRoom() string
	setRoom(room string)
	roomChangeCh() chan *roomChange
	read()
	write(msg string) error
	close()
}
