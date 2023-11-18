package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/0xPolygonID/onchain-issuer-integration-demo/server/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	auth "github.com/iden3/go-iden3-auth"
	"github.com/iden3/go-iden3-auth/loaders"
	"github.com/iden3/go-iden3-auth/pubsignals"
	"github.com/iden3/go-iden3-auth/state"
	"github.com/iden3/iden3comm/protocol"
	"github.com/patrickmn/go-cache"
	"github.com/ugorji/go/codec"
)

var (
	NgrokCallbackURL   string
	userSessionTracker = cache.New(60*time.Minute, 60*time.Minute)
	jsonHandle         codec.JsonHandle
)

type Handler struct {
	cfg config.Config
}

func NewHandler(cfg config.Config) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) GetAuthVerificationRequest(w http.ResponseWriter, r *http.Request) {
	resB, sessionId, err := h.getAuthVerificationRequest()
	if err != nil {
		log.Printf("Server -> issuer.CommHandler.GetAuthVerificationRequest() return err, err: %v", err)
		EncodeResponse(w, http.StatusInternalServerError, fmt.Sprintf("can't get auth verification request. err: %v", err))
		return
	}
	w.Header().Set("Access-Control-Expose-Headers", "x-id")
	w.Header().Set("x-id", sessionId)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	EncodeByteResponse(w, http.StatusOK, resB)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")
	tokenBytes, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Server.callback() error reading request body, err: %v", err)
		EncodeResponse(w, http.StatusBadRequest, fmt.Errorf("can't read request body"))
		return
	}

	resB, err := h.callback(sessionID, tokenBytes)
	if err != nil {
		log.Printf("Server.callback() return err, err: %v", err)
		EncodeResponse(w, http.StatusInternalServerError, fmt.Errorf("can't handle callback request"))
		return
	}

	EncodeByteResponse(w, http.StatusOK, resB)
}

func (h *Handler) GetRequestStatus(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		log.Println("Server.getRequestStatus() url parameter has invalid values")
		EncodeResponse(w, http.StatusBadRequest, fmt.Errorf("url parameter has invalid values"))
		return
	}

	resB, err := h.getRequestStatus(id)
	if err != nil {
		log.Printf("Server -> issuer.CommHandler.GetRequestStatus() return err, err: %v", err)
		EncodeResponse(w, http.StatusInternalServerError, fmt.Sprintf("can't get request status. err: %v", err))
		return
	}

	if resB == nil {
		EncodeResponse(w, http.StatusNotFound, fmt.Errorf("can't get request status with id: %s", id))
		return
	}

	EncodeByteResponse(w, http.StatusOK, resB)
}

func (h *Handler) getAuthVerificationRequest() ([]byte, string, error) {
	log.Println("Communication.GetAuthVerificationRequest() invoked")

	sId := strconv.Itoa(rand.Intn(1000000))
	uri := fmt.Sprintf("%s/api/v1/callback?sessionId=%s", NgrokCallbackURL, sId)

	request := auth.CreateAuthorizationRequestWithMessage("test flow", "message to sign", h.cfg.OnchainIssuerIdentity, uri)

	request.ID = uuid.New().String()
	request.ThreadID = uuid.New().String()

	userSessionTracker.Set(sId, request, cache.DefaultExpiration)

	msgBytes, err := json.Marshal(request)
	if err != nil {
		return nil, "", fmt.Errorf("error marshalizing response: %v", err)
	}

	return msgBytes, sId, nil
}

func (h *Handler) callback(sId string, tokenBytes []byte) ([]byte, error) {
	log.Println("Communication.Callback() invoked")

	authRequest, wasFound := userSessionTracker.Get(sId)
	if !wasFound {
		return nil, fmt.Errorf("auth request was not found for session ID: %s", sId)
	}

	var resolvers = make(map[string]pubsignals.StateResolver)
	for network, settings := range h.cfg.Resolvers {
		resolvers[network] = state.ETHResolver{
			RPCUrl:          settings.NetworkURL,
			ContractAddress: common.HexToAddress(settings.ContractState),
		}
	}
	var verificationKeyLoader = &loaders.FSKeyLoader{Dir: h.cfg.KeyDir}
	verifier := auth.NewVerifier(verificationKeyLoader, loaders.DefaultSchemaLoader{}, resolvers)
	if verifier == nil {
		return nil, fmt.Errorf("error creating verifier")
	}

	arm, err := verifier.FullVerify(context.Background(), string(tokenBytes), authRequest.(protocol.AuthorizationRequestMessage))
	if err != nil { // the verification result is false
		return nil, err
	}

	m := make(map[string]interface{})
	m["id"] = arm.From

	mBytes, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("error marshalizing response: %v", err)
	}

	userSessionTracker.Set(sId, m, cache.DefaultExpiration)

	return mBytes, nil
}

func (h *Handler) getRequestStatus(id string) ([]byte, error) {
	log.Println("Communication.Callback() invoked")

	item, ok := userSessionTracker.Get(id)
	if !ok {
		log.Printf("item not found %v", id)
		return nil, nil
	}

	switch item.(type) {
	case protocol.AuthorizationRequestMessage:
		log.Println("no authorization response yet - no data available for this request")
		return nil, nil
	case map[string]interface{}:
		b, err := json.Marshal(item)
		if err != nil {
			return nil, fmt.Errorf("error marshalizing response: %v", err)
		}
		return b, nil
	}

	return nil, fmt.Errorf("unknown item return from tracker (type %T)", item)
}

func EncodeByteResponse(w http.ResponseWriter, statusCode int, res []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, err := w.Write(res)
	if err != nil {
		log.Panicln(err)
	}
}

func EncodeResponse(w http.ResponseWriter, statusCode int, res interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := codec.NewEncoder(w, &jsonHandle).Encode(res); err != nil {
		log.Println(err)
	}
}
