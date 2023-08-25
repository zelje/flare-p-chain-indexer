package cronjob

import (
	"context"
	"flare-indexer/database"
	"flare-indexer/utils/merkle"
	"math/big"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/vms/platformvm/api"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
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
	if privateKey[:2] == "0x" {
		privateKey = privateKey[2:]
	}

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

// Deduplicate txs by txID. This is necessary because the same tx can have
// multiple UTXO inputs.
func dedupeTxs(txs []database.PChainTxData) []database.PChainTxData {
	txSet := make(map[string]*database.PChainTxData, len(txs))

	for i := range txs {
		tx := &txs[i]
		if tx.TxID == nil {
			continue
		}

		txSet[*tx.TxID] = tx
	}

	dedupedTxs := make([]database.PChainTxData, 0, len(txSet))
	for _, tx := range txSet {
		dedupedTxs = append(dedupedTxs, *tx)
	}

	return dedupedTxs
}

func buildTree(txs []database.PChainTxData) (merkle.Tree, error) {
	hashes := make([]common.Hash, len(txs))

	for i := range txs {
		tx := &txs[i]

		if tx.TxID == nil {
			return merkle.Tree{}, errors.New("tx.TxID is nil")
		}

		txHash, err := ids.FromString(*tx.TxID)
		if err != nil {
			return merkle.Tree{}, errors.Wrap(err, "ids.FromString")
		}

		hashes[i] = common.Hash(txHash)
	}

	return merkle.Build(hashes, false), nil
}

func getMerkleRoot(votingData []database.PChainTxData) (common.Hash, error) {
	tree, err := buildTree(votingData)
	if err != nil {
		return [32]byte{}, err
	}
	hash, err := tree.GetHash(0)
	if err != nil {
		return [32]byte{}, err
	}
	return hash, nil
}
