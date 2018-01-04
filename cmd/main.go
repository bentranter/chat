package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/bentranter/chat"
)

const cfgFilename = "config.toml"

var (
	tcpPortAddr   = flag.String("tcp", "3000", "tcp port")
	tcpsPortAddr  = flag.String("tcps", "3001", "secure tcp port")
	ipAddr        = flag.String("ip", "localhost", "ip address")
	logFile       = flag.String("log", "stdout", "log filename")
	httpPortAddr  = flag.String("http", "8000", "http port")
	httpsPortAddr = flag.String("https", "8001", "https port")
)

func main() {
	flag.Parse()

	cfg, err := getConfig(cfgFilename)
	if err != nil {
		log.Printf("Failed to read config: %s. Falling back to defaults.\n", err.Error())
	}
	if cfg.TCPPortAddr == "" {
		cfg.TCPPortAddr = *tcpPortAddr
	}
	if cfg.TCPSPortAddr == "" {
		cfg.TCPSPortAddr = *tcpsPortAddr
	}
	if cfg.HTTPPortAddr == "" {
		cfg.HTTPPortAddr = *httpPortAddr
	}
	if cfg.HTTPSPortAddr == "" {
		cfg.HTTPSPortAddr = *httpsPortAddr
	}
	if cfg.LogFilename == "" {
		cfg.LogFilename = *logFile
	}
	if cfg.IPAddr == "" {
		cfg.IPAddr = *ipAddr
	}

	logger := getLogger(cfg.LogFilename)
	chat.ListenAndServe(logger, cfg)
}

func getConfig(dir string) (*chat.Config, error) {
	cfg := &chat.Config{}
	_, err := os.Stat(dir)
	if err != nil {
		return cfg, err
	}

	_, err = toml.DecodeFile(dir, &cfg)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

func getLogger(filename string) *log.Logger {
	prefix := "Chat: "
	flag := log.Lshortfile | log.Ldate
	if filename == "stdout" {
		return log.New(os.Stdout, prefix, flag)
	}

	out, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil || filename == "" {
		log.Printf("Coudln't open file for logging: %s. Falling back to stdout.\n", err.Error())
		return log.New(os.Stdout, prefix, flag)
	}

	w := io.MultiWriter(out, os.Stdout)
	return log.New(w, prefix, flag)
}
