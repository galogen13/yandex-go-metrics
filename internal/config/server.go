package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Host            string `env:"ADDRESS"`
	LogLevel        string `env:"LOG_LEVEL"`
	StoreInterval   *int   `env:"STORE_INTERVAL"` // указатель, т.к. в переменной может быть 0, что важно для нас
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	RestoreStorage  *bool  `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
	StoreOnUpdate   bool
}

func GetServerConfig() (ServerConfig, error) {

	var cfg ServerConfig

	err := env.Parse(&cfg)
	if err != nil {
		return ServerConfig{}, err
	}

	hostAddressFlag := flag.String("a", "localhost:8080", "host address")
	logLevelFlag := flag.String("l", "info", "log level")
	StoreIntervalFlag := flag.Int("i", 300, "store to file interval, seconds")
	FileStoragePathFlag := flag.String("f", "./metricsstorage", "file storage path")
	DatabaseDSNFlag := flag.String("d", "host=localhost port=5432 user=postgres password=12345 dbname=metrics sslmode=disable", "file storage path")
	RestoreFlag := flag.Bool("r", false, "restore storage from file")
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

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = *DatabaseDSNFlag
	}

	if cfg.RestoreStorage == nil {
		cfg.RestoreStorage = RestoreFlag
	}

	cfg.StoreOnUpdate = (*cfg.StoreInterval == 0)

	return cfg, nil
}
