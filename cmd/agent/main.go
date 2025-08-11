package main

import (
	"flag"

	"github.com/galogen13/yandex-go-metrics/internal/agent"
)

func main() {
	hostAddress := flag.String("a", "localhost:8080", "host address")
	reportInterval := flag.Int("r", 10, "report interval, seconds")
	pollInterval := flag.Int("p", 2, "poll interval, seconds")
	flag.Parse()

	agent.Start(*hostAddress, *reportInterval, *pollInterval)
}
