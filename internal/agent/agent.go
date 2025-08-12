package agent

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/galogen13/yandex-go-metrics/internal/config"
	models "github.com/galogen13/yandex-go-metrics/internal/model"

	"github.com/go-resty/resty/v2"
)

const (
	contentTypeTextPlain = "text/plain"
)

type Agent struct {
	metrics agentMetrics
	host    string
}

type agentMetrics struct {
	// runtime.MemStats
	Alloc         float64
	BuckHashSys   float64
	Frees         float64
	GCCPUFraction float64
	GCSys         float64
	HeapAlloc     float64
	HeapIdle      float64
	HeapInuse     float64
	HeapObjects   float64
	HeapReleased  float64
	HeapSys       float64
	LastGC        float64
	Lookups       float64
	MCacheInuse   float64
	MCacheSys     float64
	MSpanInuse    float64
	MSpanSys      float64
	Mallocs       float64
	NextGC        float64
	NumForcedGC   float64
	NumGC         float64
	OtherSys      float64
	PauseTotalNs  float64
	StackInuse    float64
	StackSys      float64
	Sys           float64
	TotalAlloc    float64

	// additional
	PollCount   int64
	RandomValue float64
}

func Start(config config.AgentConfig) {

	agent := NewAgent(config.Host)

	tickerPoll := time.NewTicker(time.Duration(config.PollInterval) * time.Second)
	tickerReport := time.NewTicker(time.Duration(config.ReportInterval) * time.Second)
	for {
		select {
		case <-tickerPoll.C:
			agent.updateMetrics()
		case <-tickerReport.C:
			go agent.sendMetrics()
		}
	}
}

func NewAgent(hostAddr string) *Agent {
	return &Agent{host: hostAddr, metrics: agentMetrics{}}
}

func (agent *Agent) updateMetrics() {
	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)
	agent.metrics.Alloc = float64(rtm.Alloc)
	agent.metrics.BuckHashSys = float64(rtm.BuckHashSys)
	agent.metrics.Frees = float64(rtm.Frees)
	agent.metrics.GCCPUFraction = float64(rtm.GCCPUFraction)
	agent.metrics.GCSys = float64(rtm.GCSys)
	agent.metrics.HeapAlloc = float64(rtm.HeapAlloc)
	agent.metrics.HeapIdle = float64(rtm.HeapIdle)
	agent.metrics.HeapInuse = float64(rtm.HeapInuse)
	agent.metrics.HeapObjects = float64(rtm.HeapObjects)
	agent.metrics.HeapReleased = float64(rtm.HeapReleased)
	agent.metrics.HeapSys = float64(rtm.HeapSys)
	agent.metrics.LastGC = float64(rtm.LastGC)
	agent.metrics.Lookups = float64(rtm.Lookups)
	agent.metrics.MCacheInuse = float64(rtm.MCacheInuse)
	agent.metrics.MCacheSys = float64(rtm.MCacheSys)
	agent.metrics.MSpanInuse = float64(rtm.MSpanInuse)
	agent.metrics.MSpanSys = float64(rtm.MSpanSys)
	agent.metrics.Mallocs = float64(rtm.Mallocs)
	agent.metrics.NextGC = float64(rtm.NextGC)
	agent.metrics.NumForcedGC = float64(rtm.NumForcedGC)
	agent.metrics.NumGC = float64(rtm.NumGC)
	agent.metrics.OtherSys = float64(rtm.OtherSys)
	agent.metrics.PauseTotalNs = float64(rtm.PauseTotalNs)
	agent.metrics.StackInuse = float64(rtm.StackInuse)
	agent.metrics.StackSys = float64(rtm.StackSys)
	agent.metrics.Sys = float64(rtm.Sys)
	agent.metrics.TotalAlloc = float64(rtm.TotalAlloc)
	agent.metrics.PollCount++
	agent.metrics.RandomValue = rand.Float64()

}

func (agent Agent) sendMetrics() {

	if agent.metrics.PollCount == 0 {
		return
	}

	client := resty.New()
	client.SetRedirectPolicy(resty.RedirectPolicyFunc(
		func(req *http.Request, via []*http.Request) error {
			req.Method = http.MethodPost
			return nil
		}))

	v := reflect.ValueOf(agent.metrics)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		switch fieldValue.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			sendMetricsHTTP(client, agent.host, models.Counter, field.Name, fieldValue.Int())
		case reflect.Float32, reflect.Float64:
			sendMetricsHTTP(client, agent.host, models.Gauge, field.Name, fieldValue.Float())
		}

	}
}

func sendMetricsHTTP(client *resty.Client, host, mType, metricsName string, value any) {

	url := fmt.Sprintf("http://%s/update/%s/%s/%v", host, mType, metricsName, value)
	resp, err := client.R().
		SetHeader("Content-Type", contentTypeTextPlain).
		Post(url)
	if err != nil {
		log.Println(err)
		return
	}

	if resp.StatusCode() != http.StatusOK {
		log.Println(resp.StatusCode())
	}

}
