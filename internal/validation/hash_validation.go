package validation

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"go.uber.org/zap"
)

const (
	hashHeaderKey = "HashSHA256"
)

type HashWriter struct {
	w          http.ResponseWriter
	key        string
	body       *bytes.Buffer
	statusCode int
	headers    http.Header
}

func NewHashWriter(w http.ResponseWriter, key string) *HashWriter {
	return &HashWriter{
		w:       w,
		key:     key,
		body:    &bytes.Buffer{},
		headers: w.Header().Clone(),
	}
}

func (h *HashWriter) Header() http.Header {
	return h.headers
}

func (h *HashWriter) Write(p []byte) (int, error) {
	return h.body.Write(p)
}

func (h *HashWriter) WriteHeader(statusCode int) {
	h.statusCode = statusCode
}

func (h *HashWriter) Flush() error {
	if h.body.Len() > 0 {
		bytes, err := io.ReadAll(h.body)
		if err != nil {
			return err
		}
		_, err = h.body.Write(bytes)
		if err != nil {
			return err
		}
		hash := CalculateHMAC(bytes, h.key)
		if hash != "" {
			h.headers.Set(hashHeaderKey, hash)
		}
	}

	for key, values := range h.headers {
		for _, value := range values {
			h.w.Header().Set(key, value)
		}
	}

	if h.statusCode == 0 {
		h.statusCode = http.StatusOK
	}
	h.w.WriteHeader(h.statusCode)

	_, err := h.w.Write(h.body.Bytes())
	return err
}

func CalculateHMAC(data []byte, key string) string {
	if key == "" {
		return ""
	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func HashValidation(key string, next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		receivedHash := r.Header.Get(hashHeaderKey)
		if key == "" || receivedHash == "" {
			next.ServeHTTP(w, r)
		} else {

			body, err := readRequestBody(r)
			if err != nil {
				logger.Log.Error("unexpected error reading request body", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			expectedHash := CalculateHMAC(body, key)

			hashEquals := hmac.Equal([]byte(receivedHash), []byte(expectedHash))
			logger.Log.Info("hash check result", zap.Bool("equals", hashEquals), zap.String("key", key))

			if !hashEquals {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			hw := NewHashWriter(w, key)
			next.ServeHTTP(hw, r)
			err = hw.Flush()
			if err != nil {
				logger.Log.Error("unexpected error reading request body", zap.Error(err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}
}

func readRequestBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return []byte{}, nil
	}

	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return []byte{}, err
	}

	req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))

	return bodyBytes, nil
}
