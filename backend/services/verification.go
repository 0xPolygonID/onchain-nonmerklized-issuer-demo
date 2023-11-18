package services

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/iden3/contracts-abi/state/go/abi"
	"github.com/iden3/go-circuits/v2"
	core "github.com/iden3/go-iden3-core/v2"
	"github.com/iden3/iden3comm/v2/packers"
	"github.com/pkg/errors"
)

func StateVerificationHandler(
	ethStateContracts map[string]*abi.State,
) packers.VerificationHandlerFunc {
	return func(id circuits.CircuitID, pubsignals []string) error {

		switch id {
		case circuits.AuthV2CircuitID:
			return authV2CircuitStateVerification(ethStateContracts, pubsignals)
		default:
			return errors.Errorf("'%s' unknow circuit ID", id)
		}

	}
}

// authV2CircuitStateVerification `authV2` circuit state verification
func authV2CircuitStateVerification(
	contracts map[string]*abi.State,
	pubsignals []string,
) error {
	bytePubsig, err := json.Marshal(pubsignals)
	if err != nil {
		return err
	}

	authPubSignals := circuits.AuthV2PubSignals{}
	err = authPubSignals.PubSignalsUnmarshal(bytePubsig)
	if err != nil {
		return err
	}

	blockchain, err := core.BlockchainFromID(*authPubSignals.UserID)
	if err != nil {
		return err
	}
	networkid, err := core.NetworkIDFromID(*authPubSignals.UserID)
	if err != nil {
		return err
	}

	contract, ok := contracts[fmt.Sprintf("%s:%s", blockchain, networkid)]
	if !ok {
		return errors.Errorf("not supported blockchain prefix %s:%s", blockchain, networkid)
	}

	globalState := authPubSignals.GISTRoot.BigInt()
	globalStateInfo, err := contract.GetGISTRootInfo(&bind.CallOpts{}, globalState)
	if err != nil {
		return err
	}
	if (big.NewInt(0)).Cmp(globalStateInfo.CreatedAtTimestamp) == 0 {
		return errors.Errorf("root %s doesn't exist in smart contract", globalState.String())
	}
	if globalState.Cmp(globalStateInfo.Root) != 0 {
		return errors.Errorf("invalid global state info in the smart contract, expected root %s, got %s", globalState.String(), globalStateInfo.Root.String())
	}

	if (big.NewInt(0)).Cmp(globalStateInfo.ReplacedByRoot) != 0 && time.Since(time.Unix(globalStateInfo.ReplacedAtTimestamp.Int64(), 0)) > time.Minute*15 {
		return errors.Errorf("global state is too old, replaced timestamp is %v", globalStateInfo.ReplacedAtTimestamp.Int64())
	}

	return nil
}
