package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/iden3/go-service-template/pkg/logger"
	"github.com/iden3/go-service-template/pkg/services/issuer"
)

type IssuerHandlers struct {
	issuerService *issuer.IssuerService
}

func NewIssuerHandlers(issuerService *issuer.IssuerService) IssuerHandlers {
	return IssuerHandlers{
		issuerService: issuerService,
	}
}

func (h *IssuerHandlers) GetIssuersList(w http.ResponseWriter, r *http.Request) {
	issuers := h.issuerService.GetIssuersList(r.Context())
	w.Header().Set("Content-Type", "application/ld+json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(issuers); err != nil {
		logger.WithContext(r.Context()).WithError(err).
			Error("error marshalizing response")
	}
}
