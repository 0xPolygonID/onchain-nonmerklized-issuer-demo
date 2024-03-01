package handlers

import (
	"net/http"

	"github.com/iden3/go-service-template/pkg/logger"
	"github.com/iden3/go-service-template/pkg/services/system"
)

type SystemHandler struct {
	readinessService *system.ReadinessService
	livenessService  *system.LivenessService
}

func NewSystemHandler(
	readinessService *system.ReadinessService,
	livenessService *system.LivenessService,
) SystemHandler {
	return SystemHandler{
		readinessService: readinessService,
		livenessService:  livenessService,
	}
}

func (sh *SystemHandler) Readiness(w http.ResponseWriter, _ *http.Request) {
	status := http.StatusOK
	if !sh.readinessService.IsReady() {
		status = http.StatusServiceUnavailable
	}
	w.WriteHeader(status)
	if _, err := w.Write([]byte("OK")); err != nil {
		logger.WithError(err).Error("failed to write response")
	}
}

func (sh *SystemHandler) Liveness(w http.ResponseWriter, _ *http.Request) {
	status := http.StatusOK
	if !sh.livenessService.IsLive() {
		status = http.StatusServiceUnavailable
	}
	w.WriteHeader(status)
	if _, err := w.Write([]byte("OK")); err != nil {
		logger.WithError(err).Error("failed to write response")
	}
}
