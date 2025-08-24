package agent

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"runtime"
	"time"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"

	"github.com/go-resty/resty/v2"
)

const (
	contentTypeTextPlain = "text/plain"
	pollCounterName      = "PollCount"
)

type Agent struct {
	metrics   []metrics.Metric
	config    config.AgentConfig
	PollCount int64
}

func (agent Agent) metricIsPollCounter(name string) bool {
	return name == pollCounterName
}

func (agent *Agent) increasePollСounter() {
	agent.PollCount++
}

func (agent *Agent) resetPollCounter() {
	agent.PollCount = 0
}

func Start(config config.AgentConfig) {

	agent := NewAgent(config)

	tickerPoll := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	tickerReport := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	for {
		select {
		case <-tickerPoll.C:
			agent.updateMetrics()
		case <-tickerReport.C:
			agent.sendMetrics()
		}
	}
}

func NewAgent(agentConfig config.AgentConfig) *Agent {
	return &Agent{config: agentConfig, metrics: []metrics.Metric{}, PollCount: 0}
}

func (agent *Agent) updateMetrics() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	agent.metrics = []metrics.Metric{}

	err := agent.addNewGaugeMetric("Alloc", float64(rtm.Alloc))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("BuckHashSys", float64(rtm.BuckHashSys))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("Frees", float64(rtm.Frees))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("GCCPUFraction", float64(rtm.GCCPUFraction))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("GCSys", float64(rtm.GCSys))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("HeapAlloc", float64(rtm.HeapAlloc))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("HeapIdle", float64(rtm.HeapIdle))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("HeapInuse", float64(rtm.HeapInuse))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("HeapObjects", float64(rtm.HeapObjects))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("HeapReleased", float64(rtm.HeapReleased))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("HeapSys", float64(rtm.HeapSys))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("LastGC", float64(rtm.LastGC))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("Lookups", float64(rtm.Lookups))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("MCacheInuse", float64(rtm.MCacheInuse))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("MCacheSys", float64(rtm.MCacheSys))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("MSpanInuse", float64(rtm.MSpanInuse))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("MSpanSys", float64(rtm.MSpanSys))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("Mallocs", float64(rtm.Mallocs))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("NextGC", float64(rtm.NextGC))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("NumForcedGC", float64(rtm.NumForcedGC))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("NumGC", float64(rtm.NumGC))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("OtherSys", float64(rtm.OtherSys))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("PauseTotalNs", float64(rtm.PauseTotalNs))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("StackInuse", float64(rtm.StackInuse))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("StackSys", float64(rtm.StackSys))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("Sys", float64(rtm.Sys))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("TotalAlloc", float64(rtm.TotalAlloc))
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	err = agent.addNewGaugeMetric("RandomValue", rand.Float64())
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

	agent.increasePollСounter()
	err = agent.addNewCounterMetric(pollCounterName, agent.PollCount)
	if err != nil {
		log.Printf("error updating agent metric values: %v", err)
	}

}

func (agent *Agent) addNewGaugeMetric(mID string, value float64) error {
	metric := metrics.NewMetrics(mID, metrics.Gauge)
	if err := metric.UpdateValue(value); err != nil {
		return fmt.Errorf("error adding new gauge metric ID: %s, mType: %s, value: %v, err: %w", metric.ID, metric.MType, value, err)
	}
	agent.metrics = append(agent.metrics, metric)
	return nil
}

func (agent *Agent) addNewCounterMetric(mID string, value int64) error {
	metric := metrics.NewMetrics(mID, metrics.Counter)
	if err := metric.UpdateValue(value); err != nil {
		return fmt.Errorf("error adding new counter metric ID: %s, mType: %s, value: %v, err: %w", metric.ID, metric.MType, value, err)
	}
	agent.metrics = append(agent.metrics, metric)
	return nil
}

func (agent *Agent) sendMetrics() {

	if len(agent.metrics) == 0 {
		return
	}

	client := resty.New()
	client.SetRedirectPolicy(resty.RedirectPolicyFunc(
		func(req *http.Request, via []*http.Request) error {
			req.Method = http.MethodPost
			return nil
		}))

	for _, metric := range agent.metrics {
		err := sendMetricsHTTP(client, agent.config.Host, metric)
		if err != nil {
			log.Println(err)
			continue
		}
		if agent.metricIsPollCounter(metric.ID) {
			agent.resetPollCounter()
		}
	}

}

func sendMetricsHTTP(client *resty.Client, host string, metric metrics.Metric) error {

	url := fmt.Sprintf("http://%s/update/%s/%s/%v", host, metric.MType, metric.ID, metric.GetValue())
	resp, err := client.R().
		SetHeader("Content-Type", contentTypeTextPlain).
		Post(url)
	if err != nil {
		return fmt.Errorf("error sending POST request to url %s: %w", url, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected code while executing request to url %s: %d", url, resp.StatusCode())
	}

	return nil

}
