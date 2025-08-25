package config

import (
	"flag"
	"os"
)

type ServerConfig struct {
	Host     string
	LogLevel string
}

func GetServerConfig() ServerConfig {

	hostAddressFlag := flag.String("a", "localhost:8080", "host address")
	logLevelFlag := flag.String("l", "info", "log level")
	flag.Parse()

	hostAddress := os.Getenv("ADDRESS")
	if hostAddress == "" {
		hostAddress = *hostAddressFlag
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = *logLevelFlag
	}

	return ServerConfig{Host: hostAddress, LogLevel: logLevel}
}
