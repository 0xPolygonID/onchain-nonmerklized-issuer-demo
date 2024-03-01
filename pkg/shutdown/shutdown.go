package shutdown

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/iden3/go-service-template/pkg/logger"
)

type Shutdown interface {
	Shutdown(context.Context) error
}

type Manager struct {
	closeTimeout time.Duration
	services     []Shutdown
}

type Option func(*Manager)

func WithCloseTimeout(timeout time.Duration) Option {
	return func(s *Manager) {
		s.closeTimeout = timeout
	}
}

func NewManager(opts ...Option) *Manager {
	m := &Manager{
		closeTimeout: 5 * time.Second,
	}
	for _, opt := range opts {
		opt(m)
	}
	return &Manager{}
}

func (m *Manager) shutdown(ctx context.Context) {
	var wg sync.WaitGroup
	for _, service := range m.services {
		wg.Add(1)
		go func(s Shutdown) {
			defer wg.Done()
			if err := s.Shutdown(ctx); err != nil {
				logger.WithError(err).Error("Error during shutdown")
			}
		}(service)
	}
	wg.Wait()
}

func (m *Manager) Register(service Shutdown) {
	m.services = append(m.services, service)
}

func (m *Manager) HandleShutdownSignal() {
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan
	logger.Info("Shutting down event received")
	m.shutdown(context.Background())
	logger.Info("Shutdown completed")
}
