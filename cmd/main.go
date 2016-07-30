package main

import (
	"github.com/bentranter/torbit"
)

var config = torbit.GetConfig()
var logger = torbit.GetLogger(config.LogFilename)

func main() {
	logger := torbit.GetLogger(config.LogFilename)
	torbit.ServeTCP(logger, ":"+config.TCPPortAddr)
}
