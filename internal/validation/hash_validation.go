package validation

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
)

const (
	hashHeaderKey = "HashSHA256"
)

type hashWriter struct {
	w    http.ResponseWriter
	body []byte
	key  string
}

func newHashWriter(w http.ResponseWriter, key string) *hashWriter {
	return &hashWriter{
		w:   w,
		key: key,
	}
}

func (h *hashWriter) Header() http.Header {
	return h.w.Header()
}

func (h *hashWriter) Write(p []byte) (int, error) {
	h.body = p
	return h.w.Write(p)
}

func (c *hashWriter) WriteHeader(statusCode int) {
	hash := CalculateHMAC(c.body, c.key)
	if hash != "" {
		c.w.Header().Add(hashHeaderKey, hash)
	}

	c.w.WriteHeader(statusCode)
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
		hw := w
		if key != "" {
			body, err := readRequestBody(r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			receivedHash := r.Header.Get(hashHeaderKey)

			expectedHash := CalculateHMAC(body, key)

			if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			hw = newHashWriter(w, key)
		}

		next.ServeHTTP(hw, r)
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
