package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type AgentConfig struct {
	Host           string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
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

	return cfg, nil

}
