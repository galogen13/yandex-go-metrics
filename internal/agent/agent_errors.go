package agent

import (
	"errors"
	"net/http"
	"syscall"
)

type AgentErrorClassification int

const (
	NonRetriable AgentErrorClassification = iota
	Retriable
	Success
)

type AgentErrorClassifier struct{}

func NewAgentErrorClassifier() *AgentErrorClassifier {
	return &AgentErrorClassifier{}
}

func (c *AgentErrorClassifier) Classify(err error, statusCode int) (classification AgentErrorClassification) {
	if err != nil {
		return c.classifyError(err)
	}

	if statusCode != http.StatusOK {
		return c.classifyStatusCode(statusCode)
	}

	return Success
}

func (c *AgentErrorClassifier) classifyError(err error) AgentErrorClassification {
	if err == nil {
		return NonRetriable
	}

	var reqErr syscall.Errno
	if errors.As(err, &reqErr) {
		return classifySyscallError(reqErr)
	}

	return NonRetriable
}

func classifySyscallError(reqErr syscall.Errno) AgentErrorClassification {

	switch reqErr {
	//unix
	case syscall.ECONNREFUSED,
		syscall.ECONNRESET,
		syscall.ETIMEDOUT,
		syscall.EAGAIN:
		return Retriable

		// //windows
		// case windows.WSAECONNREFUSED,
		// 	windows.WSAECONNRESET,
		// 	windows.WSAETIMEDOUT:
		// 	return Retriable

	}

	return NonRetriable
}

func (c *AgentErrorClassifier) classifyStatusCode(statusCode int) AgentErrorClassification {

	if statusCode >= http.StatusInternalServerError ||
		statusCode == http.StatusTooManyRequests ||
		statusCode == http.StatusRequestTimeout {
		return Retriable
	}

	return NonRetriable

}
