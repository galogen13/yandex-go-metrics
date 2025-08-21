package config

import (
	"flag"
	"os"
)

type ServerConfig struct {
	Host string
}

func GetServerConfig() ServerConfig {

	hostAddress := os.Getenv("ADDRESS")

	if hostAddress == "" {
		hostAddressFlag := flag.String("a", "localhost:8080", "host address")
		flag.Parse()
		hostAddress = *hostAddressFlag
	}
	return ServerConfig{Host: hostAddress}
}
