package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	logger "github.com/iden3/go-service-template/pkg/logger"
)

// RequestLog log information about http request
func RequestLog(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		t1 := time.Now()
		defer func() {
			logger.WithContext(r.Context()).Info("http request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remoteAddr", r.RemoteAddr),
				slog.String("responseTime", fmt.Sprintf("%d ms", time.Since(t1).Milliseconds())),
				slog.Int("status", ww.Status()))
		}()

		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}
