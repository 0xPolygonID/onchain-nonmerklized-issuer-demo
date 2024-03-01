package http

import (
	"fmt"
	"log"
	"time"
)

type Option func(*Server)

func WithHost(address, port string) Option {
	return func(s *Server) {
		s.driver.Addr = fmt.Sprintf("%s:%s", address, port)
	}
}

func WithReadTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.driver.ReadTimeout = timeout
	}
}

func WithWriteTimeout(timeout time.Duration) Option {
	return func(s *Server) {
		s.driver.WriteTimeout = timeout
	}
}

func WithLogger(logger *log.Logger) Option {
	return func(s *Server) {
		s.driver.ErrorLog = logger
	}
}
