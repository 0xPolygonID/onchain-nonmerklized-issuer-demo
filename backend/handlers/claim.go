package handlers

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"

	"github.com/0xPolygonID/onchain-nonmerklized-issuer-demo/backend/common"
	"github.com/0xPolygonID/onchain-nonmerklized-issuer-demo/backend/services"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
)

type Handlers struct {
	CredentialService *services.ClaimService
	Packager          *iden3comm.PackageManager
}

type CredentialDataRequestBody struct {
	CredentialID   string    `json:"credentialId"`
	SchemaJSONLD   string    `json:"schemaJsonLd"`
	SchemaURLJSON  string    `json:"schemaUrlJson"`
	CredentialType string    `json:"credentialType"`
	Claim          [8]string `json:"claim"`
}

func (c *CredentialDataRequestBody) toCredentialData() services.CredentialData {
	var ints [8]*big.Int
	for i := range c.Claim {
		ints[i], _ = big.NewInt(0).SetString(c.Claim[i], 10)
	}
	return services.CredentialData{
		CredentialID:   c.CredentialID,
		SchemaJSONLD:   c.SchemaJSONLD,
		SchemaURLJSON:  c.SchemaURLJSON,
		CredentialType: c.CredentialType,
		Claim:          ints,
	}
}

func (h *Handlers) CreateClaim(w http.ResponseWriter, r *http.Request) {
	issuer := chi.URLParam(r, "identifier")

	issuerDID, err := w3c.ParseDID(issuer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var body CredentialDataRequestBody
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	credentialID, err := h.CredentialService.CreateVerifiableCredential(
		r.Context(),
		issuerDID,
		body.toCredentialData(),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"id": credentialID,
	})
}

func (oc *Handlers) GetUserVCByID(w http.ResponseWriter, r *http.Request) {
	issuer := chi.URLParam(r, "identifier")
	claimId := chi.URLParam(r, "credentialId")
	if claimId == "" {
		http.Error(w, "claimId query param is required", http.StatusBadRequest)
		return
	}

	w3cCredential, err := oc.CredentialService.GetCredentialByID(
		r.Context(), issuer, claimId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(w3cCredential)
}

func (h *Handlers) GetOffer(w http.ResponseWriter, r *http.Request) {
	issuer := chi.URLParam(r, "identifier")
	claimId := r.URL.Query().Get("claimId")
	if claimId == "" {
		http.Error(w, "claimId query param is required", http.StatusBadRequest)
		return
	}
	subject := r.URL.Query().Get("subject")
	if subject == "" {
		http.Error(w, "subject query param is required", http.StatusBadRequest)
		return
	}

	offerMessage := protocol.CredentialsOfferMessage{
		ID:       uuid.New().String(),
		ThreadID: uuid.New().String(),
		Typ:      packers.MediaTypePlainMessage,
		Type:     protocol.CredentialOfferMessageType,
		Body: protocol.CredentialsOfferMessageBody{
			URL: fmt.Sprintf("%s/api/v1/agent", strings.Trim(common.ExternalServerHost, "/")),
			Credentials: []protocol.CredentialOffer{
				{
					ID:          claimId,
					Description: "BalanceCredential",
				},
			},
		},
		From: issuer,
		To:   subject,
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(offerMessage)
}
