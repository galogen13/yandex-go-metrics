package agent

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"time"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"go.uber.org/zap"

	"github.com/go-resty/resty/v2"
)

const (
	contentTypeTextPlain = "text/plain"
	pollCounterName      = "PollCount"
)

type Agent struct {
	metrics   []*metrics.Metric
	config    config.AgentConfig
	PollCount int64
}

func (agent *Agent) metricIsPollCounter(name string) bool {
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

	logger.Log.Info("starting agent",
		zap.String("Host", config.Host),
		zap.String("APIFormat", config.APIFormat),
		zap.Int("PollInterval", config.PollInterval),
		zap.Any("ReportInterval", config.ReportInterval),
	)

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
	return &Agent{config: agentConfig, metrics: []*metrics.Metric{}, PollCount: 0}
}

func (agent *Agent) updateMetrics() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	agent.metrics = []*metrics.Metric{}

	err := agent.addNewGaugeMetric("Alloc", float64(rtm.Alloc))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("BuckHashSys", float64(rtm.BuckHashSys))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("Frees", float64(rtm.Frees))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("GCCPUFraction", float64(rtm.GCCPUFraction))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("GCSys", float64(rtm.GCSys))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("HeapAlloc", float64(rtm.HeapAlloc))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("HeapIdle", float64(rtm.HeapIdle))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("HeapInuse", float64(rtm.HeapInuse))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("HeapObjects", float64(rtm.HeapObjects))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("HeapReleased", float64(rtm.HeapReleased))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("HeapSys", float64(rtm.HeapSys))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("LastGC", float64(rtm.LastGC))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("Lookups", float64(rtm.Lookups))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("MCacheInuse", float64(rtm.MCacheInuse))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("MCacheSys", float64(rtm.MCacheSys))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("MSpanInuse", float64(rtm.MSpanInuse))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("MSpanSys", float64(rtm.MSpanSys))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("Mallocs", float64(rtm.Mallocs))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("NextGC", float64(rtm.NextGC))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("NumForcedGC", float64(rtm.NumForcedGC))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("NumGC", float64(rtm.NumGC))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("OtherSys", float64(rtm.OtherSys))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("PauseTotalNs", float64(rtm.PauseTotalNs))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("StackInuse", float64(rtm.StackInuse))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("StackSys", float64(rtm.StackSys))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("Sys", float64(rtm.Sys))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("TotalAlloc", float64(rtm.TotalAlloc))
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	err = agent.addNewGaugeMetric("RandomValue", rand.Float64())
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	}

	agent.increasePollСounter()
	err = agent.addNewCounterMetric(pollCounterName, agent.PollCount)
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
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
		func(req *http.Request, _ []*http.Request) error {
			req.Method = http.MethodPost
			return nil
		}))

	if agent.config.APIFormat == config.APIFormatJSONBatch {
		err := sendMetricsBatchWithJSONBody(client, agent.config.Host, agent.metrics)
		if err != nil {
			logger.Log.Error("error when send metrics batch", zap.Error(err))
			return
		}
		agent.resetPollCounter()
	} else {
		for _, metric := range agent.metrics {
			var err error
			switch agent.config.APIFormat {
			case config.APIFormatJSON:
				err = sendMetricWithJSONBody(client, agent.config.Host, metric)
			case config.APIFormatURL:
				err = sendMetricsViaPathParams(client, agent.config.Host, metric)
			}
			if err != nil {
				logger.Log.Error("error when send metric", zap.Error(err))
				continue
			}
			if agent.metricIsPollCounter(metric.ID) {
				agent.resetPollCounter()
			}
		}
	}

}

func sendMetricsViaPathParams(client *resty.Client, host string, metric *metrics.Metric) error {

	baseURL := &url.URL{
		Scheme: "http",
		Host:   host,
		Path:   path.Join("update", metric.MType, metric.ID, metric.GetValueString()),
	}
	fullURL := baseURL.String()

	resp, err := client.R().
		SetHeader("Content-Type", contentTypeTextPlain).
		Post(fullURL)

	if err != nil {
		return fmt.Errorf("error sending POST request via path params to url %s: %w", fullURL, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected code while executing request via path params to url %s: %d", fullURL, resp.StatusCode())
	}

	return nil

}

func sendMetricWithJSONBody(client *resty.Client, host string, metric *metrics.Metric) error {

	logger.Log.Info("prepairing to send metric",
		zap.String("ID", metric.ID),
		zap.String("MType", metric.MType),
		zap.Any("value", metric.GetValue()),
	)

	bodyBytes, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("error while marshalling metric with ID %s: %w", metric.ID, err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err = gz.Write(bodyBytes)
	if err != nil {
		return fmt.Errorf("error while compressing metric with ID %s: %w", metric.ID, err)
	}

	err = gz.Close()
	if err != nil {
		return fmt.Errorf("error while close compressing metric with ID %s: %w", metric.ID, err)
	}

	baseURL := &url.URL{
		Scheme: "http",
		Host:   host,
		Path:   "update",
	}
	fullURL := baseURL.String()

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf.Bytes()).
		Post(fullURL)
	if err != nil {
		return fmt.Errorf("error sending POST request with JSON body to url %s: %w", fullURL, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected code while executing request with JSON body to url %s: %d", fullURL, resp.StatusCode())
	}

	return nil

}

func sendMetricsBatchWithJSONBody(client *resty.Client, host string, metrics []*metrics.Metric) error {

	logger.Log.Info("prepairing to send metrics batch",
		zap.Any("metrics", metrics),
	)

	bodyBytes, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("error while marshalling metrics: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	_, err = gz.Write(bodyBytes)
	if err != nil {
		return fmt.Errorf("error while compressing metrics: %w", err)
	}

	err = gz.Close()
	if err != nil {
		return fmt.Errorf("error while close compressing metrics: %w", err)
	}

	baseURL := &url.URL{
		Scheme: "http",
		Host:   host,
		Path:   "updates",
	}
	fullURL := baseURL.String()

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf.Bytes()).
		Post(fullURL)
	if err != nil {
		return fmt.Errorf("error sending POST request with JSON body to url %s: %w", fullURL, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected code while executing request with JSON body to url %s: %d", fullURL, resp.StatusCode())
	}

	return nil

}
