package agent

import (
	"errors"
	"syscall"

	"golang.org/x/sys/windows"
)

type AgentErrorClassification int

const (
	NonRetriable AgentErrorClassification = iota
	Retriable
)

type AgentErrorClassifier struct{}

func NewAgentErrorClassifier() *AgentErrorClassifier {
	return &AgentErrorClassifier{}
}

func (c *AgentErrorClassifier) ClassifyError(err error) AgentErrorClassification {
	if err == nil {
		return NonRetriable
	}

	var reqErr syscall.Errno
	if errors.As(err, &reqErr) {
		return СlassifySyscallError(reqErr)
	}

	return NonRetriable
}

func (c *AgentErrorClassifier) ClassifyStatusCode(statusCode int) AgentErrorClassification {

	if statusCode >= 500 ||
		statusCode == 429 ||
		statusCode == 408 {
		return Retriable
	}

	return NonRetriable

}

func СlassifySyscallError(reqErr syscall.Errno) AgentErrorClassification {

	switch reqErr {
	//unix
	case syscall.ECONNREFUSED,
		syscall.ECONNRESET,
		syscall.ETIMEDOUT,
		syscall.EAGAIN:
		return Retriable

	//windows
	case windows.WSAECONNREFUSED,
		windows.WSAECONNRESET,
		windows.WSAETIMEDOUT:
		return Retriable

	}

	return NonRetriable
}
