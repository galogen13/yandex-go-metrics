package main

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"reflect"
	"runtime"
	"strings"
	"time"

	models "github.com/galogen13/yandex-go-metrics/internal/model"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second

	host                 = "localhost:8080/"
	contentTypeTextPlain = "text/plain"
)

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

func main() {

	metrics := agentMetrics{}

	tickerPoll := time.NewTicker(pollInterval)
	tickerReport := time.NewTicker(reportInterval)
	var rtm runtime.MemStats
	for {
		select {
		case <-tickerPoll.C:
			runtime.ReadMemStats(&rtm)
			metrics.Alloc = float64(rtm.Alloc)
			metrics.BuckHashSys = float64(rtm.BuckHashSys)
			metrics.Frees = float64(rtm.Frees)
			metrics.GCCPUFraction = float64(rtm.GCCPUFraction)
			metrics.GCSys = float64(rtm.GCSys)
			metrics.HeapAlloc = float64(rtm.HeapAlloc)
			metrics.HeapIdle = float64(rtm.HeapIdle)
			metrics.HeapInuse = float64(rtm.HeapInuse)
			metrics.HeapObjects = float64(rtm.HeapObjects)
			metrics.HeapReleased = float64(rtm.HeapReleased)
			metrics.HeapSys = float64(rtm.HeapSys)
			metrics.LastGC = float64(rtm.LastGC)
			metrics.Lookups = float64(rtm.Lookups)
			metrics.MCacheInuse = float64(rtm.MCacheInuse)
			metrics.MCacheSys = float64(rtm.MCacheSys)
			metrics.MSpanInuse = float64(rtm.MSpanInuse)
			metrics.MSpanSys = float64(rtm.MSpanSys)
			metrics.Mallocs = float64(rtm.Mallocs)
			metrics.NextGC = float64(rtm.NextGC)
			metrics.NumForcedGC = float64(rtm.NumForcedGC)
			metrics.NumGC = float64(rtm.NumGC)
			metrics.OtherSys = float64(rtm.OtherSys)
			metrics.PauseTotalNs = float64(rtm.PauseTotalNs)
			metrics.StackInuse = float64(rtm.StackInuse)
			metrics.StackSys = float64(rtm.StackSys)
			metrics.Sys = float64(rtm.Sys)
			metrics.TotalAlloc = float64(rtm.TotalAlloc)
			metrics.PollCount++
			metrics.RandomValue = rand.Float64()
		case <-tickerReport.C:

			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					req.Method = http.MethodPost
					return nil
				},
			}

			v := reflect.ValueOf(metrics)
			t := v.Type()

			for i := 0; i < v.NumField(); i++ {
				field := t.Field(i)
				fieldValue := v.Field(i)

				switch fieldValue.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					newFunction(client, models.Counter, field.Name, fieldValue.Int())
				case reflect.Float32, reflect.Float64:
					newFunction(client, models.Gauge, field.Name, fieldValue.Float())
				}

			}
		}

	}
}

func newFunction(client *http.Client, mType string, metricsName string, value any) {
	url := fmt.Sprintf("http://%s/update/%s/%s/%v", host, mType, metricsName, value)
	resp, err := client.Post(url, contentTypeTextPlain, strings.NewReader(""))

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp.StatusCode)
}
