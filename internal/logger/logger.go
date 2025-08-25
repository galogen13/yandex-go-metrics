package logger

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger = zap.NewNop()

func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = lvl
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}

type (
	responseData struct {
		status int
		size   int
	}

	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)

func RequestLogger(next http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			Log.Info("error read request body",
				zap.Error(err))
		}

		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		next(&lw, r)

		duration := time.Since(start)

		Log.Info("incoming request",
			zap.String("uri", r.RequestURI),
			zap.String("method", r.Method),
			zap.Any("body", string(bodyBytes)),
			zap.String("duration", duration.String()),
			zap.Int("status", responseData.status),
			zap.Int("size", responseData.size),
		)

	}
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.responseData.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
