package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/galogen13/yandex-go-metrics/internal/compression"
	"github.com/galogen13/yandex-go-metrics/internal/config"
	"github.com/galogen13/yandex-go-metrics/internal/crypto"
	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/retry"
	"github.com/galogen13/yandex-go-metrics/internal/service/metrics"
	"github.com/galogen13/yandex-go-metrics/internal/validation"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type HTTPConnector struct {
	host      string
	encryptor *crypto.Encryptor
	key       string
	client    *resty.Client
}

func NewHTTPConnector(config config.AgentConfig) (*HTTPConnector, error) {

	encryptor, err := crypto.NewEncryptor(config.CryptoKeyPath)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize encryptor: %w", err)
	}

	client := resty.New()
	client.SetRedirectPolicy(resty.RedirectPolicyFunc(
		func(req *http.Request, _ []*http.Request) error {
			req.Method = http.MethodPost
			return nil
		}))

	return &HTTPConnector{
			host:      config.Host,
			key:       config.Key,
			encryptor: encryptor,
			client:    client,
		},
		nil
}

func (c HTTPConnector) SendMetrics(ctx context.Context, metrics []metrics.Metric, localIP string) error {
	bodyBytes, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("error while marshalling metrics: %w", err)
	}

	compressed, err := compression.GzipCompress(bodyBytes)
	if err != nil {
		return fmt.Errorf("error while gzip compress metrics: %w", err)
	}

	body, err := c.encryptor.Encrypt(compressed.Bytes())
	if err != nil {
		return fmt.Errorf("failed to encrypt data: %w", err)
	}

	req := c.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Content-Encoding", "gzip").
		SetHeader("X-Real-IP", localIP).
		SetBody(body)

	if c.key != "" {
		hash := validation.CalculateHMAC(body, c.key)
		req.SetHeader("HashSHA256", hash)
	}

	baseURL := &url.URL{
		Scheme: "http",
		Host:   c.host,
		Path:   "updates",
	}
	fullURL := baseURL.String()

	resp, err := retry.DoWithResult(
		ctx,
		func() (*resty.Response, error) {
			return req.Post(fullURL)
		},
		NewHTTPConnectorErrorClassifier())

	if err != nil {
		return fmt.Errorf("error sending metrics via HTTP: %w", err)
	}

	logger.Log.Info("data sent", zap.String("url", fullURL), zap.Int("respCode", resp.StatusCode()))

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	return nil
}

func (c *HTTPConnector) Close() {
	logger.Log.Info("HTTP connector closed")
}
