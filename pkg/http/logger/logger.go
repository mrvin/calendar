package logger

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mrvin/calendar/internal/logger"
)

type LoggingResponseWriter struct {
	http.ResponseWriter

	statusCode    int
	totalWritByte int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{w, http.StatusOK, 0}
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LoggingResponseWriter) Write(slByte []byte) (int, error) {
	writeByte, err := lrw.ResponseWriter.Write(slByte)
	lrw.totalWritByte += writeByte

	return writeByte, err //nolint:wrapcheck
}

type Logger struct {
	Inner http.Handler
}

func (l *Logger) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	requestID := uuid.New().String()
	ctx := logger.WithRequestID(req.Context(), requestID)

	logReq := slog.With(
		slog.String("requestID", requestID),
		slog.String("method", req.Method),
		slog.String("path", req.URL.Path),
		slog.String("addr", req.RemoteAddr),
	)
	timeStart := time.Now()
	lrw := NewLoggingResponseWriter(res)
	defer func() {
		logReq.Info("Request "+req.Proto,
			slog.Int("status", lrw.statusCode),
			slog.Int("bytes", lrw.totalWritByte),
			slog.String("duration", time.Since(timeStart).String()),
		)
	}()

	l.Inner.ServeHTTP(lrw, req.WithContext(ctx))
}
