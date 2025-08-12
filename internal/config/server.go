package config

import "flag"

type ServerConfig struct {
	Host string
}

func GetServerConfig() ServerConfig {
	hostAddress := flag.String("a", "localhost:8080", "host address")

	flag.Parse()

	return ServerConfig{Host: *hostAddress}

}
