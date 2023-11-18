package parser

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	config "github.com/0xPolygonID/onchain-nonmerklized-issuer-demo/backend/common"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	identitybase "github.com/iden3/contracts-abi/identity-base/go/abi"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	"github.com/iden3/go-merkletree-sql/v2"
	"github.com/iden3/go-schema-processor/v2/merklize"
	"github.com/iden3/go-schema-processor/v2/verifiable"
	"github.com/piprate/json-gold/ld"
)

const (
	w3cBasicContext         = "https://www.w3.org/2018/credentials/v1"
	iden3proofsBasicContext = "https://schema.iden3.io/core/jsonld/iden3proofs.jsonld"
)

func NewW3Ccredential(
	c [8]*big.Int,
	issuer *w3c.DID,
	id string,
	schemaURLJSONLD string,
	schemaURLJSON string,
	credentialType string,
) (*verifiable.W3CCredential, error) {
	binInts, err := bigIntsToBytes(c)
	if err != nil {
		return nil, err
	}

	var coreClaim core.Claim
	err = coreClaim.UnmarshalBinary(binInts)
	if err != nil {
		return nil, err
	}

	exp, ok := coreClaim.GetExpirationDate()
	if !ok {
		return nil, errors.New("no expiration date")
	}

	cs, err := buildCredentialStatus(issuer, coreClaim)
	if err != nil {
		return nil, err
	}

	credentialSubject, err := buildCredentualSubject(
		schemaURLJSONLD,
		credentialType,
		coreClaim,
	)
	if err != nil {
		return nil, err
	}
	mtpProof, err := buildMTPProof(
		issuer, config.NodeRPCURL, coreClaim)
	if err != nil {
		return nil, err
	}

	return &verifiable.W3CCredential{
		ID: fmt.Sprintf("urn:uuid:%s", id),
		Context: []string{
			w3cBasicContext,
			iden3proofsBasicContext,
			schemaURLJSONLD,
		},
		Type: []string{
			"VerifiableCredential",
			credentialType,
		},
		Expiration:        &exp,
		CredentialSubject: credentialSubject,
		CredentialStatus:  cs,
		Issuer:            issuer.String(),
		CredentialSchema: verifiable.CredentialSchema{
			ID:   schemaURLJSON,
			Type: "JsonSchema2023",
		},
		Proof: verifiable.CredentialProofs{
			&mtpProof,
		},
	}, nil
}

func buildCredentualSubject(
	schemaURLJSONLD string,
	credentialType string,
	coreClaim core.Claim,
) (map[string]interface{}, error) {

	ldContext, err := ld.NewContext(
		nil,
		nil,
	).Parse(schemaURLJSONLD)
	if err != nil {
		return nil, err
	}
	serializationField, err := getSerializationAttrFromParsedContext(
		ldContext, credentialType)
	if err != nil {
		return nil, err
	}
	slots, err := parseSerializationAttr(serializationField)
	if err != nil {
		return nil, err
	}
	schemaBytes := []byte(`{"@context": "` + schemaURLJSONLD + `"}`)

	res, err := extractSlots(schemaBytes, credentialType, slots, coreClaim)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// TODO(illia-korotia): load schema one time and cache it
func extractSlots(
	schemaBytes []byte,
	credentialType string,
	slots slotsPaths,
	coreClaim core.Claim,
) (map[string]any, error) {
	credentialSubject := make(map[string]any)
	bigIntCoreClaim := coreClaim.RawSlotsAsInts()

	setFieldIfSlotExists := func(slotIdx int, slotName string) error {
		if slotName == "" {
			return nil
		}
		slotDataType, err := merklize.TypeFromContext(schemaBytes, fmt.Sprintf("%s.%s", credentialType, slotName))
		if err != nil {
			return err
		}

		v, err := convertSlotValue(bigIntCoreClaim[slotIdx], slotDataType)
		if err != nil {
			return err
		}
		credentialSubject[slotName] = v
		return nil
	}

	err := setFieldIfSlotExists(2, slots.indexAPath)
	if err != nil {
		return nil, err
	}
	err = setFieldIfSlotExists(3, slots.indexBPath)
	if err != nil {
		return nil, err
	}
	err = setFieldIfSlotExists(6, slots.valueAPath)
	if err != nil {
		return nil, err
	}
	err = setFieldIfSlotExists(7, slots.valueBPath)
	if err != nil {
		return nil, err
	}

	ownerID, err := coreClaim.GetID()
	if err != nil {
		return nil, err
	}
	ownerDID, err := core.ParseDIDFromID(ownerID)
	if err != nil {
		return nil, err
	}
	credentialSubject["id"] = ownerDID.String()
	credentialSubject["type"] = "Balance"

	return credentialSubject, nil
}

func convertSlotValue(slotValue *big.Int, datatype string) (any, error) {
	var v any
	switch datatype {
	case ld.XSDBoolean:
		if slotValue.Cmp(big.NewInt(0)) == 0 {
			v = false
		} else {
			v = true
		}
	case ld.XSDInteger:
		v = slotValue.Int64()
	case ld.XSDNS + "positiveInteger",
		ld.XSDNS + "nonNegativeInteger":
		// encode as string
		v = slotValue.String()
	/*
		case ld.XSDNS + "negativeInteger", ld.XSDNS + "nonPositiveInteger"::
		should be converted to negative int (in string representation)
		from big positive int
	*/
	default:
		return nil, fmt.Errorf("unsupported type: %v", datatype)
	}
	return v, nil
}

func buildCredentialStatus(issuerDID *w3c.DID, core core.Claim) (verifiable.CredentialStatus, error) {
	contractAddress, err := extractAddress(issuerDID)
	if err != nil {
		return verifiable.CredentialStatus{}, err
	}
	return verifiable.CredentialStatus{
		ID: fmt.Sprintf(
			"%s/credentialStatus?revocationNonce=%d&contractAddress=80001:%s",
			issuerDID.String(),
			core.GetRevocationNonce(),
			contractAddress,
		),
		Type:            verifiable.Iden3OnchainSparseMerkleTreeProof2023,
		RevocationNonce: core.GetRevocationNonce(),
	}, nil
}

func extractAddress(did *w3c.DID) (string, error) {
	id, err := core.IDFromDID(*did)
	if err != nil {
		return "", err
	}
	ca, err := core.EthAddressFromID(id)
	if err != nil {
		return "", err
	}
	return common.BytesToAddress(ca[:]).Hex(), nil
}

func bigIntsToBytes(bigs [8]*big.Int) ([]byte, error) {
	var result []byte
	for _, b := range bigs {
		binInt, err := core.NewElemBytesFromInt(b)
		if err != nil {
			return nil, err
		}
		result = append(result, binInt[:]...)
	}
	return result, nil
}

// ask Oleg about make this function public
func getSerializationAttrFromParsedContext(ldCtx *ld.Context,
	tp string) (string, error) {

	termDef, ok := ldCtx.AsMap()["termDefinitions"]
	if !ok {
		return "", errors.New("types now found in context")
	}

	termDefM, ok := termDef.(map[string]any)
	if !ok {
		return "", errors.New("terms definitions is not of correct type")
	}

	for typeName, typeDef := range termDefM {
		typeDefM, ok := typeDef.(map[string]any)
		if !ok {
			// not a type
			continue
		}
		typeCtx, ok := typeDefM["@context"]
		if !ok {
			// not a type
			continue
		}
		typeCtxM, ok := typeCtx.(map[string]any)
		if !ok {
			return "", errors.New("type @context is not of correct type")
		}
		typeID, _ := typeDefM["@id"].(string)
		if typeName != tp && typeID != tp {
			continue
		}

		serStr, _ := typeCtxM["iden3_serialization"].(string)
		return serStr, nil
	}

	return "", nil
}

type slotsPaths struct {
	indexAPath string
	indexBPath string
	valueAPath string
	valueBPath string
}

func parseSerializationAttr(serAttr string) (slotsPaths, error) {
	prefix := "iden3:v1:"
	if !strings.HasPrefix(serAttr, prefix) {
		return slotsPaths{},
			errors.New("serialization attribute does not have correct prefix")
	}
	parts := strings.Split(serAttr[len(prefix):], "&")
	if len(parts) > 4 {
		return slotsPaths{},
			errors.New("serialization attribute has too many parts")
	}
	var paths slotsPaths
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) != 2 {
			return slotsPaths{}, errors.New(
				"serialization attribute part does not have correct format")
		}
		switch kv[0] {
		case "slotIndexA":
			paths.indexAPath = kv[1]
		case "slotIndexB":
			paths.indexBPath = kv[1]
		case "slotValueA":
			paths.valueAPath = kv[1]
		case "slotValueB":
			paths.valueBPath = kv[1]
		default:
			return slotsPaths{},
				errors.New("unknown serialization attribute slot")
		}
	}
	return paths, nil
}

func buildMTPProof(
	issuerDID *w3c.DID,
	nodeURL string,
	coreClaim core.Claim,
) (verifiable.Iden3SparseMerkleTreeProof, error) {

	id, err := core.IDFromDID(*issuerDID)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}
	binAddress, err := core.EthAddressFromID(id)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	ethcli, err := ethclient.Dial(nodeURL)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	onChainIssuer, err := identitybase.NewIdentityBase(
		common.BytesToAddress(binAddress[:]),
		ethcli,
	)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	hindex, err := coreClaim.HIndex()
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	proof, err := onChainIssuer.GetClaimProof(&bind.CallOpts{}, hindex)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}
	bigState, err := onChainIssuer.GetLatestPublishedState(&bind.CallOpts{})
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}
	roots, err := onChainIssuer.GetRootsByState(&bind.CallOpts{}, bigState)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	state, err := merkletree.NewHashFromBigInt(bigState)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}
	rootOfRoots, err := merkletree.NewHashFromBigInt(roots.RootsRoot)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}
	claimTreeRoot, err := merkletree.NewHashFromBigInt(roots.ClaimsRoot)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}
	revocationTreeRoot, err := merkletree.NewHashFromBigInt(roots.RevocationsRoot)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	mtp, err := convertChainProofToMerkleProof(&proof)
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	coreClaimHex, err := coreClaim.Hex()
	if err != nil {
		return verifiable.Iden3SparseMerkleTreeProof{}, err
	}

	return verifiable.Iden3SparseMerkleTreeProof{
		Type:      verifiable.Iden3SparseMerkleTreeProofType,
		CoreClaim: coreClaimHex,
		IssuerData: verifiable.IssuerData{
			ID: issuerDID.String(),
			State: verifiable.State{
				RootOfRoots:        strpoint(rootOfRoots.Hex()),
				ClaimsTreeRoot:     strpoint(claimTreeRoot.Hex()),
				RevocationTreeRoot: strpoint(revocationTreeRoot.Hex()),
				Value:              strpoint(state.Hex()),
			},
		},
		MTP: mtp,
	}, nil
}

func strpoint(s string) *string {
	return &s
}

func convertChainProofToMerkleProof(smtProof *identitybase.SmtLibProof) (*merkletree.Proof, error) {
	var (
		existence bool
		nodeAux   *merkletree.NodeAux
		err       error
	)

	if smtProof.Existence {
		existence = true
	} else {
		existence = false
		if smtProof.AuxExistence {
			nodeAux = &merkletree.NodeAux{}
			nodeAux.Key, err = merkletree.NewHashFromBigInt(smtProof.AuxIndex)
			if err != nil {
				return nil, err
			}
			nodeAux.Value, err = merkletree.NewHashFromBigInt(smtProof.AuxValue)
			if err != nil {
				return nil, err
			}
		}
	}

	allSiblings := make([]*merkletree.Hash, len(smtProof.Siblings))
	for i, s := range smtProof.Siblings {
		sh, err := merkletree.NewHashFromBigInt(s)
		if err != nil {
			return nil, err
		}
		allSiblings[i] = sh
	}

	p, err := merkletree.NewProofFromData(existence, allSiblings, nodeAux)
	if err != nil {
		return nil, err
	}

	return p, nil
}
