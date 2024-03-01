package packagemanager

import (
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/iden3/contracts-abi/state/go/abi"
	"github.com/iden3/go-circuits/v2"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/go-jwz/v2"
	"github.com/iden3/iden3comm/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/pkg/errors"
)

type state struct {
	contracts map[int]*abi.State
}

func (s *state) verify(_ circuits.CircuitID, pubsignals []string) error {
	bytePubsig, err := json.Marshal(pubsignals)
	if err != nil {
		return errors.Errorf("error marshaling pubsignals: %v", err)
	}

	authPubSignals := circuits.AuthV2PubSignals{}
	err = authPubSignals.PubSignalsUnmarshal(bytePubsig)
	if err != nil {
		return errors.Errorf("error unmarshaling pubsignals: %v", err)
	}

	did, err := core.ParseDIDFromID(*authPubSignals.UserID)
	if err != nil {
		return errors.Errorf("error parsing DID from ID '%s': %v",
			authPubSignals.UserID.String(), err)
	}

	chainID, err := core.ChainIDfromDID(*did)
	if err != nil {
		return errors.Errorf("error getting chainID from DID '%s': %v",
			did, err)
	}

	contract, ok := s.contracts[int(chainID)]
	if !ok {
		return errors.Errorf("not supported blockchain %d", chainID)
	}

	globalState := authPubSignals.GISTRoot.BigInt()
	globalStateInfo, err := contract.GetGISTRootInfo(&bind.CallOpts{}, globalState)
	if err != nil {
		return errors.Errorf("error getting global state info '%s': %v", globalState, err)
	}
	if (big.NewInt(0)).Cmp(globalStateInfo.CreatedAtTimestamp) == 0 {
		return errors.Errorf("root '%s' doesn't exist in smart contract", globalState)
	}
	if globalState.Cmp(globalStateInfo.Root) != 0 {
		return errors.Errorf("invalid global state info in the smart contract, expected root '%s', got '%s'",
			globalState.String(), globalStateInfo.Root.String())
	}

	if (big.NewInt(0)).Cmp(globalStateInfo.ReplacedByRoot) != 0 && time.Since(time.Unix(globalStateInfo.ReplacedAtTimestamp.Int64(), 0)) > time.Minute*15 {
		return errors.Errorf("global state is too old, replaced timestamp is %v", globalStateInfo.ReplacedAtTimestamp.Int64())
	}

	return nil
}

func NewPackageManager(
	supportedRPC map[string]string,
	supportedStateContracts map[string]string,
	circuitsFolderPath string,
) (*iden3comm.PackageManager, error) {
	authV2Path := fmt.Sprintf("%s/authV2.json", circuitsFolderPath)
	verificationKey, err := os.ReadFile(authV2Path)
	if err != nil {
		return nil, errors.Errorf(
			"issuer with the file verification_key.json by path '%s': %v", authV2Path, err)
	}

	states := state{
		contracts: make(map[int]*abi.State, len(supportedStateContracts)),
	}
	for chainID, stateAddr := range supportedStateContracts {
		rpcURL, ok := supportedRPC[chainID]
		if !ok {
			return nil, errors.Errorf("not supported RPC for blockchain '%s'", chainID)
		}
		ec, err := ethclient.Dial(rpcURL)
		if err != nil {
			return nil, errors.Errorf("error creating eth client: %v", err)
		}
		stateContract, err := abi.NewState(common.HexToAddress(stateAddr), ec)
		if err != nil {
			return nil, errors.Errorf("error creating state contract ABI: %v", err)
		}
		v, err := strconv.Atoi(chainID)
		if err != nil {
			return nil, errors.Errorf("invalid chainID '%s': %v", chainID, err)
		}
		states.contracts[v] = stateContract
	}

	verifications := make(map[jwz.ProvingMethodAlg]packers.VerificationParams)
	verifications[jwz.AuthV2Groth16Alg] = packers.NewVerificationParams(
		verificationKey,
		states.verify,
	)

	zkpPackerV2 := packers.NewZKPPacker(
		nil,
		verifications,
	)

	packageManager := iden3comm.NewPackageManager()

	err = packageManager.RegisterPackers(zkpPackerV2, &packers.PlainMessagePacker{})
	if err != nil {
		return nil, errors.Errorf("error registering packers: %v", err)
	}

	return packageManager, nil
}
