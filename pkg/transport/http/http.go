package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/iden3/go-service-template/pkg/logger"
)

type Server struct {
	driver *http.Server
}

func New(handler http.Handler, opts ...Option) *Server {
	s := &Server{
		driver: &http.Server{
			Addr:         ":8080",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			Handler:      handler,
			ErrorLog:     slog.NewLogLogger(slog.Default().Handler(), slog.LevelError),
		},
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.driver.Shutdown(ctx)
}

func (s *Server) Start() error {
	logger.Info("HTTP server started", slog.String("address", s.driver.Addr))
	return s.driver.ListenAndServe()
}
