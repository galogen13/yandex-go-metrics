package connector

import (
	"errors"
	"syscall"

	"github.com/galogen13/yandex-go-metrics/internal/classification"
	"github.com/galogen13/yandex-go-metrics/internal/retry"
)

type HTTPConnectorErrorClassifier struct{}

func NewHTTPConnectorErrorClassifier() *HTTPConnectorErrorClassifier {
	return &HTTPConnectorErrorClassifier{}
}

func (c *HTTPConnectorErrorClassifier) Classify(err error) retry.ErrorClassification {
	if err == nil {
		return retry.NonRetriable
	}

	var reqErr syscall.Errno
	if errors.As(err, &reqErr) {
		return classifySyscallError(reqErr)
	}

	return retry.NonRetriable
}

func classifySyscallError(reqErr syscall.Errno) retry.ErrorClassification {

	if classification.IsRetriableSyscallError(reqErr) {
		return retry.Retriable
	}

	return retry.NonRetriable
}
