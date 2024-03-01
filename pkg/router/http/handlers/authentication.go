package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/iden3/go-service-template/pkg/logger"
	"github.com/iden3/go-service-template/pkg/services/authentication"
)

type AuthenticationHandlers struct {
	callbackURL           string
	authenticationService *authentication.AuthenticationService
}

func NewAuthenticationHandlers(
	callbackURL string,
	authenticationService *authentication.AuthenticationService,
) AuthenticationHandlers {
	return AuthenticationHandlers{
		callbackURL:           callbackURL,
		authenticationService: authenticationService,
	}
}

func (h *AuthenticationHandlers) CreateAuthenticationRequest(w http.ResponseWriter, r *http.Request) {
	issuerDIDStr := r.URL.Query().Get("issuer")
	if issuerDIDStr == "" {
		logger.Error("issuer is required")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	uri := fmt.Sprintf("%s/api/v1/callback", h.callbackURL)
	request, sessionID := h.authenticationService.NewAuthenticationRequest(uri, issuerDIDStr)
	w.Header().Set("Access-Control-Expose-Headers", "x-id")
	w.Header().Set("x-id", sessionID)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if err := json.NewEncoder(w).Encode(request); err != nil {
		logger.WithError(err).Error("error marshalizing response", slog.Any("request", request))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *AuthenticationHandlers) Callback(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	tokenBytes, err := io.ReadAll(r.Body)
	if err != nil {
		logger.WithError(err).Error("error reading body")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID, err := h.authenticationService.Verify(r.Context(), sessionID, tokenBytes)
	if err != nil {
		logger.WithError(err).Error("error verifying token", slog.String("sessionID", sessionID))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"id": userID,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithError(err).Error("error marshalizing response", slog.Any("response", response))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *AuthenticationHandlers) AuthenticationRequestStatus(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("id")
	userID, err := h.authenticationService.AuthenticationRequestStatus(sessionID)
	if err != nil {
		logger.WithError(err).Error("error getting session", slog.String("sessionID", sessionID))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	response := map[string]string{
		"id": userID,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.WithError(err).Error("error marshalizing response", slog.Any("sessionID", sessionID))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
