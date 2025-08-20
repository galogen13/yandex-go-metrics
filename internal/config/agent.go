package config

import "flag"

type AgentConfig struct {
	Host           string
	ReportInterval int
	PollInterval   int
}

func GetAgentConfig() AgentConfig {
	hostAddress := flag.String("a", "localhost:8080", "host address")
	reportInterval := flag.Int("r", 10, "report interval, seconds")
	pollInterval := flag.Int("p", 2, "poll interval, seconds")
	flag.Parse()

	return AgentConfig{
		Host:           *hostAddress,
		ReportInterval: *reportInterval,
		PollInterval:   *pollInterval}

}
