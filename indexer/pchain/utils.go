package pchain

import (
	"context"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/utils/json"
	"github.com/ava-labs/avalanchego/vms/components/avax"
	"github.com/ava-labs/avalanchego/vms/platformvm/genesis"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ybbus/jsonrpc/v3"
)

func CallPChainGetTxApi(client jsonrpc.RPCClient, txID string) (*txs.Tx, error) {
	id, err := ids.FromString(txID)
	if err != nil {
		return nil, err
	}

	// In case of genesis? transaction
	if id == ids.Empty {
		return nil, nil
	}

	// Fetch from chain
	params := api.GetTxArgs{
		TxID:     id,
		Encoding: formatting.Hex,
	}
	reply := api.GetTxReply{}
	ctx := context.Background()
	response, err := client.Call(ctx, "platform.getTx", params)
	if err != nil {
		return nil, err
	}
	err = response.GetObject(&reply)
	if err != nil {
		return nil, err
	}

	// Parse from hex string
	txData, err := formatting.Decode(formatting.Hex, reply.Tx.(string))
	if err != nil {
		return nil, err
	}
	tx, err := txs.Parse(genesis.Codec, txData)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

// Copy-paste from
// 	"github.com/ava-labs/avalanchego/vms/platformvm/service"
// To avoid an additional dependency
type GetRewardUTXOsReply struct {
	NumFetched json.Uint64         `json:"numFetched"`
	UTXOs      []string            `json:"utxos"`
	Encoding   formatting.Encoding `json:"encoding"`
}

func CallPChainGetRewardUTXOsApi(client jsonrpc.RPCClient, txID string) ([]*avax.UTXO, error) {
	id, err := ids.FromString(txID)
	if err != nil {
		return nil, err
	}

	params := api.GetTxArgs{
		TxID:     id,
		Encoding: formatting.Hex,
	}
	reply := GetRewardUTXOsReply{}
	ctx := context.Background()
	response, err := client.Call(ctx, "platform.getRewardUTXOs", params)
	if err != nil {
		return nil, err
	}
	err = response.GetObject(&reply)
	if err != nil {
		return nil, err
	}

	result := []*avax.UTXO(nil)
	for _, utxoHex := range reply.UTXOs {
		txData, err := formatting.Decode(formatting.Hex, utxoHex)
		if err != nil {
			return nil, err
		}
		utxo := avax.UTXO{}
		_, err = txs.Codec.Unmarshal(txData, &utxo)
		if err != nil {
			return nil, err
		}
		result = append(result, &utxo)
	}
	return result, nil
}
