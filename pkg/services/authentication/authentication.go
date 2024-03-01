package authentication

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/google/uuid"
	auth "github.com/iden3/go-iden3-auth/v2"
	"github.com/iden3/iden3comm/v2/protocol"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
)

var (
	userSessionTracker = cache.New(60*time.Minute, 60*time.Minute)
)

type AuthenticationService struct {
	verifier *auth.Verifier
}

func NewAuthenticationService(verifier *auth.Verifier) *AuthenticationService {
	return &AuthenticationService{
		verifier: verifier,
	}
}

func (a *AuthenticationService) NewAuthenticationRequest(
	serviceURL string,
	issuer string,
) (request protocol.AuthorizationRequestMessage, sessionID string) {
	//nolint:gosec // this is not a security issue
	sessionID = strconv.Itoa(rand.Intn(1000000))
	uri := fmt.Sprintf("%s?sessionId=%s", serviceURL, sessionID)
	request = auth.CreateAuthorizationRequestWithMessage(
		"login to website", "", issuer, uri,
	)
	request.ID = uuid.New().String()
	request.ThreadID = uuid.New().String()
	userSessionTracker.Set(sessionID, request, cache.DefaultExpiration)
	return request, sessionID
}

func (a *AuthenticationService) Verify(ctx context.Context,
	sessionID string, tokenBytes []byte) (string, error) {
	request, found := userSessionTracker.Get(sessionID)
	if !found {
		return "", errors.Errorf("auth request was not found for session ID: %s", sessionID)
	}
	authResponse, err := a.verifier.FullVerify(
		ctx,
		string(tokenBytes),
		request.(protocol.AuthorizationRequestMessage),
	)
	if err != nil {
		return "", errors.Errorf("error verifying token: %v", err)
	}
	userSessionTracker.Set(sessionID, authResponse.From, cache.DefaultExpiration)
	return authResponse.From, nil
}

func (a *AuthenticationService) AuthenticationRequestStatus(sessionID string) (string, error) {
	session, found := userSessionTracker.Get(sessionID)
	if !found {
		return "", errors.Errorf("session not found: %s", sessionID)
	}
	switch s := session.(type) {
	case string:
		return s, nil
	default:
		return "", errors.Errorf("session didn't pass authentification")
	}
}
