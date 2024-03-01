package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-service-template/pkg/logger"
	"github.com/iden3/go-service-template/pkg/router/http/middleware"
	"github.com/iden3/go-service-template/pkg/services/issuer"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
)

type convertClaimRequest struct {
	HexData string `json:"hexData"`
	Version string `json:"version"`
}

type IssuerHandlers struct {
	issuerService *issuer.IssuerService
	host          string
}

func NewIssuerHandlers(issuerService *issuer.IssuerService, host string) IssuerHandlers {
	return IssuerHandlers{
		issuerService: issuerService,
		host:          host,
	}
}

func (h *IssuerHandlers) ConvertClaim(w http.ResponseWriter, r *http.Request) {
	convertRequest := &convertClaimRequest{}
	if err := json.NewDecoder(r.Body).Decode(convertRequest); err != nil {
		logger.WithContext(r.Context()).WithError(err).
			Error("error unmarshalizing request")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	issuerDID := r.Context().Value(middleware.DIDContextKey{}).(*w3c.DID)
	recordID, err := h.issuerService.ConvertHexDataToVerifiableCredential(
		r.Context(),
		issuerDID,
		convertRequest.HexData,
		convertRequest.Version,
	)
	if err != nil {
		logger.WithContext(r.Context()).WithError(err).
			Error("error issuing credential")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(w).Encode(map[string]string{"id": recordID}); err != nil {
		logger.WithContext(r.Context()).WithError(err).
			Error("error marshalizing response")
	}
}

func (h *IssuerHandlers) GetOffer(w http.ResponseWriter, r *http.Request) {
	claimID := r.URL.Query().Get("claimId")
	if claimID == "" {
		logger.WithContext(r.Context()).Error("claimId query param is required")
		http.Error(w, "claimId query param is required", http.StatusBadRequest)
		return
	}
	issuerDID := r.Context().Value(middleware.DIDContextKey{}).(*w3c.DID)
	subject := r.URL.Query().Get("subject")
	if subject == "" {
		logger.WithContext(r.Context()).Error("subject query param is required")
		http.Error(w, "subject query param is required", http.StatusBadRequest)
		return
	}

	offerMessage := protocol.CredentialsOfferMessage{
		ID:       uuid.New().String(),
		ThreadID: uuid.New().String(),
		Typ:      packers.MediaTypePlainMessage,
		Type:     protocol.CredentialOfferMessageType,
		Body: protocol.CredentialsOfferMessageBody{
			URL: fmt.Sprintf("%s/api/v1/agent", strings.Trim(h.host, "/")),
			Credentials: []protocol.CredentialOffer{
				{
					ID:          claimID,
					Description: "BalanceCredential",
				},
			},
		},
		From: issuerDID.String(),
		To:   subject,
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(offerMessage); err != nil {
		logger.WithContext(r.Context()).WithError(err).
			Error("error marshalizing response")
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
