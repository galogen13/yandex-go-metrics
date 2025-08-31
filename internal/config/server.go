package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Host            string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	StoreInterval   *int   `env:"STORE_INTERVAL"` // указатель, т.к. в переменной может быть 0, что важно для нас
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         *bool  `env:"RESTORE"`
}

func GetServerConfig() (ServerConfig, error) {

	var cfg ServerConfig

	err := env.Parse(&cfg)
	if err != nil {
		return ServerConfig{}, err
	}

	fmt.Println(os.Environ())

	hostAddressFlag := flag.String("a", "localhost:8080", "host address")
	logLevelFlag := flag.String("l", "info", "log level")
	StoreIntervalFlag := flag.Int("i", 300, "store interval")
	FileStoragePathFlag := flag.String("l", "./", "file storage path")
	RestoreFlag := flag.Bool("l", false, "restore")
	flag.Parse()

	if cfg.Host == "" {
		cfg.Host = *hostAddressFlag
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = *logLevelFlag
	}

	if cfg.StoreInterval == nil {
		cfg.StoreInterval = StoreIntervalFlag
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = *FileStoragePathFlag
	}

	if cfg.Restore == nil {
		cfg.Restore = RestoreFlag
	}

	return cfg, nil
}
