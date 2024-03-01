package issuer

import (
	"context"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	ethclientlib "github.com/ethereum/go-ethereum/ethclient"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-iden3-core/v2/w3c"
	convertor "github.com/iden3/go-onchain-credential-adapter"
	"github.com/iden3/go-schema-processor/v2/merklize"
	"github.com/iden3/go-service-template/pkg/repository"
	"github.com/piprate/json-gold/ld"
	"github.com/pkg/errors"
)

type IssuerService struct {
	repository     *repository.CredentialRepository
	ethclients     map[string]*ethclientlib.Client
	issuers        []string
	documentLoader ld.DocumentLoader
}

func NewIssuerService(
	credentialRepository *repository.CredentialRepository,
	ethclients map[string]*ethclientlib.Client,
	issuers []string,
	documentLoader ld.DocumentLoader,
) *IssuerService {
	return &IssuerService{
		repository:     credentialRepository,
		ethclients:     ethclients,
		issuers:        issuers,
		documentLoader: documentLoader,
	}
}

func (is *IssuerService) GetIssuersList(_ context.Context) []string {
	return is.issuers
}

func (is *IssuerService) ConvertHexDataToVerifiableCredential(
	ctx context.Context,
	issuerDID *w3c.DID,
	hexOnchainData string,
	version string,
) (string, error) {

	ethclient, _, err := is.lookforEthConnectForDID(issuerDID)
	if err != nil {
		return "", errors.Wrapf(err, "error getting ethclient for issuer: '%s'", issuerDID)
	}

	verifiableCredential, err := convertor.W3CCredentialFromOnchainHex(
		ctx,
		ethclient,
		issuerDID,
		hexOnchainData,
		version,
		convertor.WithMerklizeOptions(merklize.Options{
			DocumentLoader: is.documentLoader,
		}),
	)
	if err != nil {
		return "", errors.Wrap(err, "error converting onchain data to verifiable credential")
	}

	id, err := is.repository.Create(ctx, verifiableCredential)
	if err != nil {
		return "", errors.Wrap(err, "error creating credential")
	}
	return id, nil
}

func (is *IssuerService) lookforEthConnectForDID(did *w3c.DID) (ethclient *ethclientlib.Client, contractAddress string, err error) {
	issuerID, err := core.IDFromDID(*did)
	if err != nil {
		return nil, "", errors.Wrap(err, "error getting issuer ID")
	}
	networkID, err := core.ChainIDfromDID(*did)
	if err != nil {
		return nil, "", errors.Wrapf(err, "network not found for did '%s'", did)
	}
	ethclient, found := is.ethclients[strconv.Itoa(int(networkID))]
	if !found {
		return nil, "", errors.Errorf("ethclient not found for network id '%d'", networkID)
	}
	contractAddressBytes, err := core.EthAddressFromID(issuerID)
	if err != nil {
		return nil, "", errors.Wrap(err, "error getting contract address")
	}
	return ethclient, common.BytesToAddress(contractAddressBytes[:]).Hex(), nil
}
