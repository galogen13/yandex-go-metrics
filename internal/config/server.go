package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type ServerConfig struct {
	Host                 string `env:"ADDRESS"`
	LogLevel             string `env:"LOG_LEVEL"`
	StoreInterval        *int   `env:"STORE_INTERVAL"` // указатель, т.к. в переменной может быть 0, что важно для нас
	FileStoragePath      string `env:"FILE_STORAGE_PATH"`
	RestoreStorage       *bool  `env:"RESTORE"`
	DatabaseDSN          string `env:"DATABASE_DSN"`
	UseDatabaseAsStorage bool
	StoreOnUpdate        bool
	StorePeriodically    bool
}

func GetServerConfig() (*ServerConfig, error) {

	var cfg ServerConfig

	err := env.Parse(&cfg)
	if err != nil {
		return nil, err
	}

	hostAddressFlag := flag.String("a", "localhost:8080", "host address")
	logLevelFlag := flag.String("l", "info", "log level")
	storeIntervalFlag := flag.Int("i", 300, "store to file interval, seconds")
	fileStoragePathFlag := flag.String("f", "./metricsstorage", "file storage path")
	databaseDSNFlag := flag.String("d", "", "file storage path")
	restoreFlag := flag.Bool("r", false, "restore storage from file")
	flag.Parse()

	if cfg.Host == "" {
		cfg.Host = *hostAddressFlag
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = *logLevelFlag
	}

	if cfg.StoreInterval == nil {
		cfg.StoreInterval = storeIntervalFlag
	}

	if cfg.FileStoragePath == "" {
		cfg.FileStoragePath = *fileStoragePathFlag
	}

	if cfg.RestoreStorage == nil {
		cfg.RestoreStorage = restoreFlag
	}

	if cfg.DatabaseDSN == "" {
		cfg.DatabaseDSN = *databaseDSNFlag
	}

	cfg.UseDatabaseAsStorage = (cfg.DatabaseDSN != "")

	cfg.StoreOnUpdate = (*cfg.StoreInterval == 0) && !cfg.UseDatabaseAsStorage

	cfg.StorePeriodically = (*cfg.StoreInterval != 0) && !cfg.UseDatabaseAsStorage

	*cfg.RestoreStorage = *cfg.RestoreStorage && !cfg.UseDatabaseAsStorage

	return &cfg, nil
}
