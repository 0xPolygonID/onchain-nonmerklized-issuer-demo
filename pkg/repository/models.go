package repository

import (
	"encoding/json"
	"time"

	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/pkg/errors"
)

type credentialStatusModel struct {
	ID              string
	Type            string
	RevocatioNnonce uint64
}

type credentialModel struct {
	ID                string
	Context           []string
	Type              []string
	Expiration        string
	IssuanceDate      string
	CredentialSubject map[string]interface{}
	CredentialStatus  credentialStatusModel
	Issuer            string
	CredentialSchema  verifiable.CredentialSchema
	Proof             interface{}
}

func NewCredentailModelFromW3C(vc *verifiable.W3CCredential) (credentialModel, error) {
	// since siblings filend is private on Proof we should extract
	// this filed for JSON marshaling and unmarshaling
	// alternatively we can store raw json in the database
	// and execute only one Marshal
	tmp, err := json.Marshal(vc.Proof)
	if err != nil {
		return credentialModel{}, errors.Wrapf(err, "failed to marshal proof")
	}
	var fullProof interface{}
	if err = json.Unmarshal(tmp, &fullProof); err != nil {
		return credentialModel{}, errors.Wrapf(err, "failed to unmarshal proof")
	}

	cs := vc.CredentialStatus.(*verifiable.CredentialStatus)
	return credentialModel{
		ID:                vc.ID,
		Context:           vc.Context,
		Type:              vc.Type,
		Expiration:        vc.Expiration.Format(time.RFC3339Nano),
		IssuanceDate:      vc.IssuanceDate.Format(time.RFC3339Nano),
		CredentialSubject: vc.CredentialSubject,
		CredentialStatus: credentialStatusModel{
			ID:              cs.ID,
			Type:            string(cs.Type),
			RevocatioNnonce: cs.RevocationNonce,
		},
		Issuer:           vc.Issuer,
		CredentialSchema: vc.CredentialSchema,
		Proof:            fullProof,
	}, nil
}

func (cm *credentialModel) ToW3C() (*verifiable.W3CCredential, error) {
	tmp, err := json.Marshal(cm.Proof)
	if err != nil {
		return nil,
			errors.Wrapf(err, "failed to marshal proof")
	}

	mtpProofs := []*verifiable.Iden3SparseMerkleTreeProof{}
	if err := json.Unmarshal(tmp, &mtpProofs); err != nil {
		return nil,
			errors.Wrapf(err, "failed to unmarshal proof")
	}

	proofs := make(verifiable.CredentialProofs, 0, len(mtpProofs))
	for _, proof := range mtpProofs {
		proofs = append(proofs, proof)
	}

	expTime, err := time.Parse(time.RFC3339Nano, cm.Expiration)
	if err != nil {
		return nil,
			errors.Wrapf(err, "failed to parse expiration time '%s'", cm.Expiration)
	}
	issuanceTime, err := time.Parse(time.RFC3339Nano, cm.IssuanceDate)
	if err != nil {
		return nil,
			errors.Wrapf(err, "failed to parse issuance time '%s'", cm.IssuanceDate)
	}

	return &verifiable.W3CCredential{
		ID:                cm.ID,
		Context:           cm.Context,
		Type:              cm.Type,
		Expiration:        &expTime,
		IssuanceDate:      &issuanceTime,
		CredentialSubject: cm.CredentialSubject,
		CredentialStatus: verifiable.CredentialStatus{
			ID:              cm.CredentialStatus.ID,
			Type:            verifiable.CredentialStatusType(cm.CredentialStatus.Type),
			RevocationNonce: cm.CredentialStatus.RevocatioNnonce,
		},
		Issuer:           cm.Issuer,
		CredentialSchema: cm.CredentialSchema,
		Proof:            proofs,
	}, nil
}
