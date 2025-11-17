package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/galogen13/yandex-go-metrics/internal/compression"
	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/retry"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"github.com/galogen13/yandex-go-metrics/internal/validation"
	"go.uber.org/zap"

	"github.com/go-resty/resty/v2"
)

const (
	pollCounterName = "PollCount"
)

type Agent struct {
	metrics        []*metrics.Metric
	muxMetrics     *sync.Mutex
	muxPollCounter *sync.Mutex
	config         config.AgentConfig
	PollCount      int64
}

func (agent *Agent) increasePollСounter() {
	agent.PollCount++
}

func (agent *Agent) decreasePollCounter(decrementer int64) {

	if agent.PollCount-decrementer < 0 {
		agent.PollCount = 0
	} else {
		agent.PollCount -= decrementer
	}

}

func Start(config config.AgentConfig) {

	agent := NewAgent(config)

	logger.Log.Info("starting agent",
		zap.String("Host", config.Host),
		zap.Int("PollInterval", config.PollInterval),
		zap.Any("ReportInterval", config.ReportInterval),
		zap.Int("RateLimit", config.RateLimit),
	)

	const numJobs = 1
	jobs := make(chan any, numJobs)
	for w := 1; w <= agent.config.RateLimit; w++ {
		go agent.startSendWorker(jobs)
	}

	defer close(jobs)

	tickerPoll := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	tickerReport := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	for {
		select {
		case <-tickerPoll.C:
			go agent.updateMetrics()
		case <-tickerReport.C:
			jobs <- nil
		}
	}

}

func (agent *Agent) startSendWorker(jobs <-chan any) {
	for range jobs {
		agent.sendMetrics()
	}
}

func NewAgent(agentConfig config.AgentConfig) *Agent {
	return &Agent{
		config:         agentConfig,
		metrics:        []*metrics.Metric{},
		muxMetrics:     &sync.Mutex{},
		muxPollCounter: &sync.Mutex{}}
}

func (agent *Agent) updateMetrics() {

	doneCh := make(chan struct{})
	defer close(doneCh)

	channels := fanOut(doneCh)
	addResultCh := fanIn(doneCh, channels...)

	metrics := multiplyMetrics(doneCh, addResultCh)

	agent.muxPollCounter.Lock()
	agent.increasePollСounter()
	metric, err := newCounterMetric(pollCounterName, agent.PollCount)
	agent.muxPollCounter.Unlock()
	if err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		metrics = append(metrics, metric)
	}

	agent.muxMetrics.Lock()
	defer agent.muxMetrics.Unlock()
	agent.metrics = metrics

}

type metricsResult struct {
	metrics []*metrics.Metric
}

func fanOut(doneCh chan struct{}) []chan metricsResult {

	workers := []func() []*metrics.Metric{
		getRuntimeMetrics,
		getPSMetrics,
	}
	channels := make([]chan metricsResult, len(workers))

	for i, worker := range workers {
		addResultCh := startWorker(doneCh, worker)
		channels[i] = addResultCh
	}

	return channels
}

func startWorker(doneCh chan struct{}, getMetrics func() []*metrics.Metric) chan metricsResult {
	addRes := make(chan metricsResult)

	go func() {
		defer close(addRes)

		result := metricsResult{metrics: getMetrics()}

		select {
		case <-doneCh:
			return
		case addRes <- result:
		}

	}()
	return addRes
}

func fanIn(doneCh chan struct{}, resultChs ...chan metricsResult) chan metricsResult {

	finalCh := make(chan metricsResult)

	var wg sync.WaitGroup

	for _, ch := range resultChs {

		chClosure := ch

		wg.Add(1)

		go func() {
			defer wg.Done()

			for data := range chClosure {
				select {
				case <-doneCh:
					return
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}

func multiplyMetrics(doneCh chan struct{}, inputCh chan metricsResult) []*metrics.Metric {
	result := []*metrics.Metric{}

	for data := range inputCh {

		select {
		case <-doneCh:
			return nil
		default:
			result = append(result, data.metrics...)
		}
	}

	return result
}

func getRuntimeMetrics() []*metrics.Metric {

	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	result := []*metrics.Metric{}

	if metric, err := newGaugeMetric("Alloc", float64(rtm.Alloc)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("BuckHashSys", float64(rtm.BuckHashSys)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("Frees", float64(rtm.Frees)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("GCCPUFraction", float64(rtm.GCCPUFraction)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("GCSys", float64(rtm.GCSys)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("HeapAlloc", float64(rtm.HeapAlloc)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("HeapIdle", float64(rtm.HeapIdle)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("HeapInuse", float64(rtm.HeapInuse)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("HeapObjects", float64(rtm.HeapObjects)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("HeapReleased", float64(rtm.HeapReleased)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("HeapSys", float64(rtm.HeapSys)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("LastGC", float64(rtm.LastGC)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("Lookups", float64(rtm.Lookups)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("MCacheInuse", float64(rtm.MCacheInuse)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("MCacheSys", float64(rtm.MCacheSys)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("MSpanInuse", float64(rtm.MSpanInuse)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("MSpanSys", float64(rtm.MSpanSys)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("Mallocs", float64(rtm.Mallocs)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("NextGC", float64(rtm.NextGC)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("NumForcedGC", float64(rtm.NumForcedGC)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("NumGC", float64(rtm.NumGC)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("OtherSys", float64(rtm.OtherSys)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("PauseTotalNs", float64(rtm.PauseTotalNs)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("StackInuse", float64(rtm.StackInuse)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("StackSys", float64(rtm.StackSys)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("Sys", float64(rtm.Sys)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("TotalAlloc", float64(rtm.TotalAlloc)); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	if metric, err := newGaugeMetric("RandomValue", rand.Float64()); err != nil {
		logger.Log.Error("error updating agent metric values", zap.Error(err))
	} else {
		result = append(result, metric)
	}

	return result

}

func getPSMetrics() []*metrics.Metric {

	result := []*metrics.Metric{}

	if vmStat, err := mem.VirtualMemory(); err == nil {

		if metric, err := newGaugeMetric("TotalMemory", float64(vmStat.Total)); err != nil {
			logger.Log.Error("error updating agent metric values", zap.Error(err))
		} else {
			result = append(result, metric)
		}

		if metric, err := newGaugeMetric("FreeMemory", float64(vmStat.Free)); err != nil {
			logger.Log.Error("error updating agent metric values", zap.Error(err))
		} else {
			result = append(result, metric)
		}

	} else {
		logger.Log.Error("error getting virtual memory metrics", zap.Error(err))
	}

	if cpuPercent, err := cpu.Percent(0, true); err == nil {
		for cpuNum, usage := range cpuPercent {
			if metric, err := newGaugeMetric(fmt.Sprintf("CPUutilization%d", cpuNum), float64(usage)); err != nil {
				logger.Log.Error("error updating agent metric values", zap.Error(err))
			} else {
				result = append(result, metric)
			}
		}
	} else {
		logger.Log.Error("error getting cpu metrics", zap.Error(err))
	}

	return result
}

func newGaugeMetric(mID string, value float64) (*metrics.Metric, error) {
	metric := metrics.NewMetrics(mID, metrics.Gauge)
	if err := metric.UpdateValue(value); err != nil {
		return nil, fmt.Errorf("error adding new gauge metric ID: %s, mType: %s, value: %v, err: %w", metric.ID, metric.MType, value, err)
	}
	return metric, nil
}

func newCounterMetric(mID string, value int64) (*metrics.Metric, error) {
	metric := metrics.NewMetrics(mID, metrics.Counter)
	if err := metric.UpdateValue(value); err != nil {
		return nil, fmt.Errorf("error adding new counter metric ID: %s, mType: %s, value: %v, err: %w", metric.ID, metric.MType, value, err)
	}
	return metric, nil
}

func (agent *Agent) sendMetrics() {

	agent.muxMetrics.Lock()
	if len(agent.metrics) == 0 {
		logger.Log.Info("nothing to send")
		return
	}
	currentPollCounterMetricValue := agent.metrics[len(agent.metrics)-1].GetValue() // последняя метрика - PollCounter
	currentPollCounter, err := metrics.CounterValue(currentPollCounterMetricValue)
	if err != nil {
		logger.Log.Error("cannot read currentPollCounter", zap.Error(err))
		return
	}

	logger.Log.Debug("prepairing to send metrics batch",
		zap.Any("metrics", agent.metrics),
	)

	bodyBytes, err := json.Marshal(agent.metrics)
	agent.muxMetrics.Unlock()
	if err != nil {
		logger.Log.Error("error while marshalling metrics", zap.Error(err))
		return
	}

	buf, err := compression.GzipCompress(bodyBytes)
	if err != nil {
		logger.Log.Error("error while gzip compress metrics", zap.Error(err))
		return
	}

	client := resty.New()
	client.SetRedirectPolicy(resty.RedirectPolicyFunc(
		func(req *http.Request, _ []*http.Request) error {
			req.Method = http.MethodPost
			return nil
		}))

	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetBody(buf.Bytes())

	if agent.config.Key != "" {
		hash := validation.CalculateHMAC(buf.Bytes(), agent.config.Key)
		req.SetHeader("HashSHA256", hash)
	}

	baseURL := &url.URL{
		Scheme: "http",
		Host:   agent.config.Host,
		Path:   "updates",
	}
	fullURL := baseURL.String()

	resp, err := retry.DoWithResult(
		context.Background(),
		func() (*resty.Response, error) {
			return req.Post(fullURL)
		},
		NewAgentErrorClassifier())

	if err != nil {
		logger.Log.Error("error sending metrics", zap.Error(err))
		return
	}

	logger.Log.Info("data sent", zap.String("url", fullURL), zap.Int("respCode", resp.StatusCode()))

	if resp.StatusCode() != http.StatusOK {
		logger.Log.Error("unexpected status code", zap.Int("status code", resp.StatusCode()))
		return
	}

	agent.muxPollCounter.Lock()
	defer agent.muxPollCounter.Unlock()
	agent.decreasePollCounter(currentPollCounter)

}
