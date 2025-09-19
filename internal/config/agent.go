package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

const (
	APIFormatJSON      string = "json"
	APIFormatURL       string = "url"
	APIFormatJSONBatch string = "json-batch"
)

type AgentConfig struct {
	Host           string `env:"ADDRESS"`         // адрес сервера, на который будут отправляться метрики
	ReportInterval int    `env:"REPORT_INTERVAL"` // количество секунд между отправками метрик на сервер
	PollInterval   int    `env:"POLL_INTERVAL"`   // количество секунд между сборами значений метрик

	// для поддержки старого варианта передачи значений метрик внутри URL установить значение "url",
	// "json" - для передачи каждой метрики отдельным запросом, иначе "json-batch" - передача одинм пакетом всех метрик.
	APIFormat string `env:"API_FORMAT"`
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
	apiFormat := flag.String("f", APIFormatJSONBatch, fmt.Sprintf("API format: %q, %q or %q", APIFormatJSONBatch, APIFormatJSON, APIFormatURL))
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

	if cfg.APIFormat == "" {
		cfg.APIFormat = *apiFormat
	}

	if cfg.APIFormat != APIFormatJSON && cfg.APIFormat != APIFormatURL && cfg.APIFormat != APIFormatJSONBatch {
		return AgentConfig{}, fmt.Errorf("unexpected api format: %v", cfg.APIFormat)
	}

	return cfg, nil

}
