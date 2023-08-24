package cronjob

import (
	"context"
	"math/big"

	"github.com/ava-labs/avalanchego/vms/platformvm/api"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pkg/errors"
	"github.com/ybbus/jsonrpc/v3"
)

type PermissionedValidators struct {
	Validators []*api.PermissionedValidator
}

func CallPChainGetConnectedValidators(client jsonrpc.RPCClient) ([]*api.PermissionedValidator, error) {
	ctx := context.Background()
	response, err := client.Call(ctx, "platform.getCurrentValidators")
	if err != nil {
		return nil, err
	}

	reply := PermissionedValidators{}
	err = response.GetObject(&reply)

	return reply.Validators, err
}

func TransactOptsFromPrivateKey(privateKey string, chainID int) (*bind.TransactOpts, error) {
	pk, err := crypto.HexToECDSA(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "crypto.HexToECDSA")
	}

	opts, err := bind.NewKeyedTransactorWithChainID(
		pk, big.NewInt(int64(chainID)),
	)
	if err != nil {
		return nil, errors.Wrap(err, "bind.NewKeyedTransactorWithChainID")
	}
	// bind.N
	return opts, nil
}
