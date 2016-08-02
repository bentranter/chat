package torbit

type client interface {
	getName() string
	read()
	write(msg string) error
	close()
}
