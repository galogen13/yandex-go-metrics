package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

const (
	ApiFormatJSON string = "json"
	ApiFormatURL  string = "url"
)

type AgentConfig struct {
	Host           string `env:"ADDRESS"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	PollInterval   int    `env:"POLL_INTERVAL"`
	ApiFormat      string `env:"API_FORMAT"`
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
	apiFormat := flag.String("f", ApiFormatJSON, "API format")
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

	if cfg.ApiFormat == "" {
		cfg.ApiFormat = *apiFormat
	}

	if cfg.ApiFormat != ApiFormatJSON && cfg.ApiFormat != ApiFormatURL {
		return AgentConfig{}, fmt.Errorf("unexpected api format: %v", cfg.ApiFormat)
	}

	return cfg, nil

}
