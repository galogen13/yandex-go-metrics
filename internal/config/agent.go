package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	flag "github.com/spf13/pflag"
)

type AgentConfig struct {
	Host           string `env:"ADDRESS"`         // адрес сервера, на который будут отправляться метрики
	ReportInterval int    `env:"REPORT_INTERVAL"` // количество секунд между отправками метрик на сервер
	PollInterval   int    `env:"POLL_INTERVAL"`   // количество секунд между сборами значений метрик
	Key            string `env:"KEY"`             // ключ
	RateLimit      int    `env:"RATE_LIMIT"`      // максимальное количество горутин, одновременно отправляющих данные на сервер
	CryptoKeyPath  string `env:"CRYPTO_KEY"`      // путь к публичному ключу
	ConfigFile     string `env:"CONFIG"`
}

type FileAgentConfig struct {
	Host           string `json:"address"`         // адрес сервера, на который будут отправляться метрики
	ReportInterval string `json:"report_interval"` // время между отправками метрик на сервер
	PollInterval   string `json:"poll_interval"`   // время между сборами значений метрик
	Key            string `json:"key"`             // ключ
	RateLimit      int    `json:"rate_limit"`      // максимальное количество горутин, одновременно отправляющих данные на сервер
	CryptoKeyPath  string `json:"crypto_key"`      // путь к публичному ключу
}

func GetAgentConfig() (AgentConfig, error) {

	var cfg AgentConfig

	err := env.Parse(&cfg)
	if err != nil {
		return cfg, err
	}

	hostAddress := flag.StringP("address", "a", "localhost:8080", "host address")
	reportInterval := flag.IntP("report-interval", "r", 10, "report interval, seconds")
	pollInterval := flag.IntP("poll-interval", "p", 2, "poll interval, seconds")
	key := flag.StringP("key", "k", "", "secret key")
	rateLimit := flag.IntP("rate-limit", "l", 1, "rate limit")
	cryptoKeyPath := flag.String("crypto-key", "", "crypto key path")
	configFile := flag.StringP("config", "c", "", "config file")
	flag.Parse()

	if cfg.Host == "" {
		cfg.Host = *hostAddress
	}

	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = *reportInterval
	}

	if cfg.PollInterval == 0 {
		cfg.PollInterval = *pollInterval
	}

	if cfg.Key == "" {
		cfg.Key = *key
	}

	if cfg.RateLimit == 0 {
		cfg.RateLimit = *rateLimit
	}

	if cfg.CryptoKeyPath == "" {
		cfg.CryptoKeyPath = *cryptoKeyPath
	}

	if cfg.ConfigFile == "" {
		cfg.ConfigFile = *configFile
	}

	err = cfg.parseConfigFile()
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func (cfg *AgentConfig) parseConfigFile() error {
	if cfg.ConfigFile == "" {
		return nil
	}

	file, err := os.Open(cfg.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	var fileConfig FileAgentConfig
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&fileConfig)
	if err != nil {
		return fmt.Errorf("failed to decode config file: %w", err)
	}

	if cfg.Host == "" {
		cfg.Host = fileConfig.Host
	}

	if cfg.ReportInterval == 0 {
		reportIntervalDuration, err := time.ParseDuration(fileConfig.ReportInterval)
		if err != nil {
			return fmt.Errorf("failed to parse ReportInterval duration: %w", err)
		}
		cfg.ReportInterval = int(reportIntervalDuration.Seconds())
	}

	if cfg.PollInterval == 0 {
		pollIntervalDuration, err := time.ParseDuration(fileConfig.PollInterval)
		if err != nil {
			return fmt.Errorf("failed to parse PollInterval duration: %w", err)
		}
		cfg.PollInterval = int(pollIntervalDuration.Seconds())
	}

	if cfg.Key == "" {
		cfg.Key = fileConfig.Key
	}

	if cfg.RateLimit == 0 {
		cfg.RateLimit = fileConfig.RateLimit
	}

	if cfg.CryptoKeyPath == "" {
		cfg.CryptoKeyPath = fileConfig.CryptoKeyPath
	}

	return nil
}
