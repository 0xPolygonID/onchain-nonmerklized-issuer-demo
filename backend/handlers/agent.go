package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/iden3/iden3comm/v2/protocol"
)

func (h *Handlers) Agent(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 2*1000*1000)
	envelope, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("failed read request body", err)
		http.Error(w, "can't bind request to protocol message", http.StatusBadRequest)
		return
	}

	basicMessage, err := h.Packager.UnpackWithType(packers.MediaTypeZKPMessage, envelope)
	if err != nil {
		log.Println("failed unpack protocol message", err)
		http.Error(w, "failed unpack protocol message", http.StatusBadRequest)
		return
	}

	if basicMessage.ID == "" {
		log.Println("empty 'id' field")
		http.Error(w, "empty 'id' field", http.StatusBadRequest)
		return
	}

	if basicMessage.To == "" {
		log.Println("empty 'to' field")
		http.Error(w, "empty 'to' field", http.StatusBadRequest)
		return
	}

	// WARN: need more validation as in identity server

	var (
		resp           []byte
		httpStatusCode = http.StatusOK
	)
	switch basicMessage.Type {
	case protocol.CredentialFetchRequestMessageType:
		resp, err = h.handleCredentialFetchRequest(r.Context(), basicMessage)
		if err != nil {
			log.Println("failed handling credential fetch request")
			http.Error(w, "failed handling credential fetch request", http.StatusBadRequest)
			return
		}
		fmt.Println("credential bytes", string(resp))
	default:
		log.Printf("failed handling %s status request", basicMessage.Type)
		http.Error(w,
			fmt.Sprintf("failed handling %s status request", basicMessage.Type), http.StatusBadRequest)
		return
	}

	_, err = w3c.ParseDID(basicMessage.From)
	if err != nil {
		log.Println("failed get sender from request")
		http.Error(w, "failed get sender from request", http.StatusBadRequest)
		return
	}

	var respBytes []byte
	if len(resp) > 0 {
		respBytes, err = h.Packager.Pack(packers.MediaTypePlainMessage, resp, packers.PlainPackerParams{})
		if err != nil {
			log.Println("failed pack response")
			http.Error(w, "failed create jwz token", http.StatusBadRequest)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatusCode)
	_, err = w.Write(respBytes)

}

func (h *Handlers) handleCredentialFetchRequest(ctx context.Context, basicMessage *iden3comm.BasicMessage) ([]byte, error) {
	if basicMessage.To == "" {
		return nil, errors.New("failed request. empty 'to' field")
	}

	if basicMessage.From == "" {
		return nil, errors.New("failed request. empty 'from' field")
	}

	fetchRequestBody := &protocol.CredentialFetchRequestMessageBody{}
	err := json.Unmarshal(basicMessage.Body, fetchRequestBody)
	if err != nil {
		return nil, fmt.Errorf("invalid credential fetch request body: %w", err)
	}

	issuerDID, err := w3c.ParseDID(basicMessage.To)
	if err != nil {
		return nil, fmt.Errorf("invalid issuer id in base message: %w", err)
	}

	userDID, err := w3c.ParseDID(basicMessage.From)
	if err != nil {
		return nil, fmt.Errorf("invalid user id in base message: %w", err)
	}

	var claimID uuid.UUID
	claimID, err = uuid.Parse(fetchRequestBody.ID)
	if err != nil {
		return nil, fmt.Errorf("invalid claim id in fetch request body: %w", err)
	}

	cred, err := h.CredentialService.GetCredentialByID(
		ctx, issuerDID.String(), claimID.String())
	if err != nil {
		return nil, fmt.Errorf("failed get claim by claimID: %w", err)
	}

	if cred.CredentialSubject["id"] != userDID.String() {
		return nil, errors.New("claim doesn't relate to sender")
	}

	if err != nil {
		return nil, fmt.Errorf("failed convert claim: %w", err)
	}

	resp := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(resp)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(&protocol.CredentialIssuanceMessage{
		ID:       uuid.NewString(),
		Type:     protocol.CredentialIssuanceResponseMessageType,
		ThreadID: basicMessage.ThreadID,
		Body:     protocol.IssuanceMessageBody{Credential: cred},
		From:     basicMessage.To,
		To:       basicMessage.From,
	})
	return resp.Bytes(), err
}
