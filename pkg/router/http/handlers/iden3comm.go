package handlers

import (
	"io"
	"net/http"

	"github.com/iden3/go-service-template/pkg/logger"
	"github.com/iden3/go-service-template/pkg/services/iden3comm"
)

const (
	iden3commMsgBodySize = 2 * 1000 * 1000
)

type Iden3commHandlers struct {
	iden3commService *iden3comm.Iden3commService
}

func NewIden3commHandlers(iden3commService *iden3comm.Iden3commService) Iden3commHandlers {
	return Iden3commHandlers{
		iden3commService: iden3commService,
	}
}

func (h *Iden3commHandlers) Agent(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, iden3commMsgBodySize)
	envelope, err := io.ReadAll(r.Body)
	if err != nil {
		logger.WithContext(r.Context()).WithError(err).Error("failed read request body")
		http.Error(w, "can't bind request to protocol message", http.StatusBadRequest)
		return
	}
	bytesMsg, err := h.iden3commService.Handle(r.Context(), envelope)
	if err != nil {
		logger.WithContext(r.Context()).WithError(err).Error("failed handle request")
		http.Error(w, "failed handle request", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(bytesMsg)
	if err != nil {
		logger.WithContext(r.Context()).WithError(err).Error("failed write response")
	}
}
