package pchain

import (
	"context"

	"github.com/ava-labs/avalanchego/api"
	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting"
	"github.com/ava-labs/avalanchego/vms/platformvm/genesis"
	"github.com/ava-labs/avalanchego/vms/platformvm/txs"
	"github.com/ybbus/jsonrpc/v3"
)

func CallPChainGetTxApi(client jsonrpc.RPCClient, txID string) (*txs.Tx, error) {
	id, err := ids.FromString(txID)
	if err != nil {
		return nil, err
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
