package torbit

import (
	"flag"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// this reaallllyy needs to go in main...
var cfg *Config

const cfgFilename = "config.toml"

// just populate config file based on it...?
func init() {
	initcfg := &Config{}
	flag.StringVar(&initcfg.TCPPortAddr, "tcp", "3000", "tcp service address")
	flag.StringVar(&initcfg.IPAddr, "ip", "localhost", "ip service address")
	flag.StringVar(&initcfg.LogFilename, "log", "", "log file location")
	flag.StringVar(&initcfg.HTTPPortAddr, "http", "8000", "http service address")
	flag.Parse()
	cfg = initcfg
}

type Config struct {
	TCPPortAddr  string
	IPAddr       string
	LogFilename  string
	HTTPPortAddr string
}

func GetConfig() *Config {
	usr, err := user.Current()
	if err != nil {
		println("Couldn't get user: ", err.Error())
		return cfg
	}

	_, err = os.Stat(filepath.Join(usr.HomeDir, cfgFilename))
	if err != nil {
		return cfg
	}

	_, err = toml.DecodeFile(filepath.Join(usr.HomeDir, cfgFilename), &cfg)
	if err != nil {
		println("Error decoding file: ", err.Error())
	}
	return cfg
}

func GetLogger(filename string) *log.Logger {
	prefix := "Torbit Challenge: "
	flag := log.Lshortfile
	out, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil || filename == "" {
		return log.New(os.Stdout, prefix, flag)
	}
	return log.New(out, prefix, flag)
}
