package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	Host           string `env:"ADDRESS"`         // адрес сервера, на который будут отправляться метрики
	ReportInterval int    `env:"REPORT_INTERVAL"` // количество секунд между отправками метрик на сервер
	PollInterval   int    `env:"POLL_INTERVAL"`   // количество секунд между сборами значений метрик
	Key            string `env:"KEY"`             // ключ
	RateLimit      int    `env:"RATE_LIMIT"`      // максимальное количество горутин, одновременно отправляющих данные на сервер
	CryptoKeyPath  string `env:"CRYPTO_KEY"`      // путь к публичному ключу
}

func GetAgentConfig() (AgentConfig, error) {

	var cfg AgentConfig

	err := env.Parse(&cfg)
	if err != nil {
		return cfg, err
	}

	hostAddress := flag.String("a", "localhost:8080", "host address")
	reportInterval := flag.Int("r", 10, "report interval, seconds")
	pollInterval := flag.Int("p", 2, "poll interval, seconds")
	key := flag.String("k", "", "secret key")
	rateLimit := flag.Int("l", 1, "rate limit")
	cryptoKeyPath := flag.String("crypto-key", "", "crypto key path")
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

	return cfg, nil

}
