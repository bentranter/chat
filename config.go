package torbit

import (
	"flag"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

var (
	tcpPortAddr  string
	ipAddr       string
	logFilename  string
	httpPortAddr string

	cfgFilepath = configFile()
)

const cfgFilename = "config.toml"

func init() {
	flag.StringVar(&tcpPortAddr, "tcp", "3000", "tcp service address")
	flag.StringVar(&ipAddr, "ip", "localhost", "ip service address")
	flag.StringVar(&logFilename, "log", "", "log file location")
	flag.StringVar(&httpPortAddr, "http", "8000", "http service address")
}

type Config struct {
	TCPPortAddr  string
	IPAddr       string
	LogFilename  string
	HTTPPortAddr string
}

func GetConfig() *Config {
	cfg, err := readConfig()
	if err != nil {
		// we can't read from the file, so a new one should
		// probs be made
		println(err.Error())
	}

	// flags override these
	if tcpPortAddr != "" {
		cfg.TCPPortAddr = tcpPortAddr
	}
	if ipAddr != "" {
		cfg.IPAddr = ipAddr
	}
	if logFilename != "" {
		cfg.LogFilename = logFilename
	}
	if httpPortAddr != "" {
		cfg.HTTPPortAddr = httpPortAddr
	}

	return cfg
}

func GetLogger(filename string) *log.Logger {
	prefix := "Torbit Challenge: "
	flag := log.Lshortfile

	out, err := os.OpenFile(cfgFilepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil || filename == "" {
		return log.New(os.Stdout, prefix, flag)
	}
	return log.New(out, prefix, flag)
}

func readConfig() (*Config, error) {
	cfg := &Config{}
	_, err := os.Stat(cfgFilepath)
	if err != nil {
		writeDefaultConfig()
		return cfg, err
	}
	_, err = toml.DecodeFile(filepath.Dir(cfgFilename), &cfg)
	if err != nil {
		// file must be mangled. just use defaults
		// and rebuild file
		writeDefaultConfig()
		println("cant decode, creating new file", err.Error())
		return cfg, nil
	}
	return cfg, nil
}

func writeDefaultConfig() error {
	file, err := os.OpenFile(cfgFilepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	return toml.NewEncoder(file).Encode(&Config{
		TCPPortAddr:  tcpPortAddr,
		IPAddr:       ipAddr,
		LogFilename:  logFilename,
		HTTPPortAddr: httpPortAddr,
	})
}

func configFile() string {
	usr, err := user.Current()
	if err != nil {
		return ""
	}
	return filepath.Join(usr.HomeDir, cfgFilename)
}
