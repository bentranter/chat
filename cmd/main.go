package main

import (
	"bufio"
	"net"
	"os"

	"github.com/bentranter/torbit"
)

var config = torbit.GetConfig()
var logger = torbit.GetLogger(config.LogFilename)

func handleConn(conn net.Conn, clientID int, msgs chan string) {
	// read each message and jam it in the messages channel
	r := bufio.NewReader(conn)
	for {
		in, err := r.ReadString('\n')
		if err != nil {
			// should signal that someone disconnected
			logger.Println(err.Error())
			break
		}
		// use the clientID, but they should have a name
		msgs <- string(clientID) + in
	}

	// if we've errored, we'll be out of the loop for that conn, so close the
	// connection
	conn.Close()
	logger.Println("Conn closed for: ", clientID)
}

func broadcast(conn net.Conn, msg string) {
	_, err := conn.Write([]byte(msg))
	if err != nil {
		// you chould disconnect the client here, but it's better to
		// have a disconnect channel
	}
}

func main() {
	logger := torbit.GetLogger(config.LogFilename)

	// each client gets this dumb id
	clientCount := 0

	// all the connected clients
	clients := make(map[net.Conn]int)

	// incoming connections
	newConn := make(chan net.Conn)

	// incoming messages
	msgs := make(chan string)

	server, err := net.Listen("tcp", ":"+config.TCPPortAddr)
	if err != nil {
		logger.Fatalln(err)
		os.Exit(1)
	}
	logger.Println("Server started.")

	// accept incoming connections
	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				logger.Fatalln(err)
				os.Exit(1)
			}
			newConn <- conn
		}
	}()

	for {
		select {
		case conn := <-newConn:
			clientCount++
			logger.Println("New client: ", clientCount)
			clients[conn] = clientCount

			// write client's messages
			go handleConn(conn, clients[conn], msgs)

		case msg := <-msgs:
			// log the message, don't put it in the loop obv
			logger.Print("New message: ", msg)

			// broadcast
			for conn := range clients {
				go broadcast(conn, msg)
			}
		}
	}
}
