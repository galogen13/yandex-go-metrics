package audit

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"github.com/galogen13/yandex-go-metrics/internal/retry"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

type AuditService struct {
	Auditors []Auditor
}

func NewAuditService() *AuditService {
	return &AuditService{Auditors: []Auditor{}}
}

func (as *AuditService) Register(auditor Auditor) {
	as.Auditors = append(as.Auditors, auditor)
}

func (as *AuditService) Notify(auditLog AuditLog) {
	for _, auditor := range as.Auditors {
		go auditor.Notify(auditLog)
	}
}

type AuditLog struct {
	Timestamp int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

func NewAuditLog(metricNames []string, ip string) AuditLog {
	return AuditLog{Timestamp: time.Now().Unix(), Metrics: metricNames, IPAddress: ip}
}

type Auditor interface {
	Notify(AuditLog)
}

type URLAuditor struct {
	urlStr string
	client *resty.Client
}

func (urlAuditor URLAuditor) Notify(auditLog AuditLog) {

	auditJSON, err := json.Marshal(auditLog)
	if err != nil {
		logger.Log.Error("cannot marshal audit log to json", zap.Error(err))
		return
	}

	req := urlAuditor.client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(auditJSON)

	baseURL, err := url.Parse(urlAuditor.urlStr)
	if err != nil {
		logger.Log.Error("error parsing url", zap.Error(err))
		return
	}

	fullURL := baseURL.String()

	resp, err := retry.DoWithResult(
		context.Background(),
		func() (*resty.Response, error) {
			return req.Post(fullURL)
		},
		NewAuditErrorClassifier())

	if err != nil {
		logger.Log.Error("error notify url auditor", zap.Error(err))
		return
	}

	logger.Log.Info("data sent to url auditor", zap.String("url", fullURL), zap.Int("respCode", resp.StatusCode()))

	if resp.StatusCode() != http.StatusOK {
		logger.Log.Error("unexpected status code", zap.Int("status code", resp.StatusCode()))
		return
	}

}

func NewURLAuditor(urlStr string) (*URLAuditor, error) {
	if urlStr == "" {
		return nil, errors.New("url auditor: URL not filled")
	}
	return &URLAuditor{urlStr: urlStr, client: resty.New()}, nil
}

type FileAuditor struct {
	filePath string
}

func NewFileAuditor(filePath string) (*FileAuditor, error) {
	if filePath == "" {
		return nil, errors.New("file auditor: file path not filled")
	}
	return &FileAuditor{filePath: filePath}, nil
}

func (fileAuditor FileAuditor) Notify(auditLog AuditLog) {
	file, err := os.OpenFile(fileAuditor.filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Log.Error("cannot open/create file", zap.Error(err))
		return
	}
	defer file.Close()

	auditJSON, err := json.MarshalIndent(auditLog, "	", "")
	if err != nil {
		logger.Log.Error("cannot marshal audit log to json", zap.Error(err))
		return
	}

	auditJSON = append(auditJSON, '\n')

	_, err = file.Write(auditJSON)
	if err != nil {
		logger.Log.Error("cannot write audit log to file", zap.Error(err))
		return
	}
}
