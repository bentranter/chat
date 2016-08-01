package torbit

type client interface {
	getID() uint64
	getName() string
	setName(name string)
	read()
	write(msg string) error
	close()
}
