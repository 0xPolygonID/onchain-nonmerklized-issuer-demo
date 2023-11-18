package services

import (
	"context"
	"math/big"

	parser "github.com/0xPolygonID/onchain-nonmerklized-issuer-demo/backend/pkg"
	"github.com/0xPolygonID/onchain-nonmerklized-issuer-demo/backend/repository"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-schema-processor/v2/verifiable"
)

type ClaimService struct {
	CredentialRepository *repository.CredentialRepository
}

type CredentialData struct {
	CredentialID   string
	SchemaJSONLD   string
	SchemaURLJSON  string
	CredentialType string
	Claim          [8]*big.Int
}

func (cs *ClaimService) CreateVerifiableCredential(
	ctx context.Context,
	issuer *w3c.DID,
	credentialData CredentialData,
) (string, error) {
	parser, err := parser.NewW3Ccredential(
		credentialData.Claim,
		issuer,
		credentialData.CredentialID,
		credentialData.SchemaJSONLD,
		credentialData.SchemaURLJSON,
		credentialData.CredentialType,
	)
	if err != nil {
		return "", err
	}

	id, err := cs.CredentialRepository.Create(ctx, *parser)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (cs *ClaimService) GetCredentialByID(
	ctx context.Context,
	issuer string,
	credentialID string,
) (verifiable.W3CCredential, error) {
	return cs.CredentialRepository.GetVCByID(
		ctx,
		issuer,
		credentialID,
	)
}
