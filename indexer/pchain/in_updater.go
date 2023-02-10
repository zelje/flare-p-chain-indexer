package pchain

import (
	"flare-indexer/database"
	"flare-indexer/indexer/context"
	"flare-indexer/indexer/shared"
	"flare-indexer/utils"

	"github.com/ybbus/jsonrpc/v3"
	"gorm.io/gorm"
)

type pChainInputUpdater struct {
	shared.BaseInputUpdater

	db     *gorm.DB
	client jsonrpc.RPCClient
}

func newPChainInputUpdater(ctx context.IndexerContext, client jsonrpc.RPCClient) *pChainInputUpdater {
	ioUpdater := pChainInputUpdater{
		db:     ctx.DB(),
		client: client,
	}
	ioUpdater.InitCache(ctx.Config().Indexer.OutputsCacheSize)
	return &ioUpdater
}

func (iu *pChainInputUpdater) UpdateInputs(inputs map[string][]shared.Input) error {
	err := iu.UpdateInputsFromCache(inputs)
	if err != nil {
		return err
	}
	err = iu.updateFromDB(inputs)
	if err != nil {
		return err
	}
	return iu.updateFromChain(inputs)
}

// notUpdated is a map from *output* id to inputs referring this output
func (iu *pChainInputUpdater) updateFromDB(notUpdated map[string][]shared.Input) error {
	outs, err := database.FetchPChainTxOutputs(iu.db, utils.Keys(notUpdated))
	if err != nil {
		return err
	}
	baseOuts := make([]shared.Output, len(outs))
	for i, o := range outs {
		baseOuts[i] = &o.TxOutput
	}
	return shared.UpdateInputsWithOutputs(notUpdated, baseOuts)
}

// notUpdated is a map from *output* id to inputs referring this output
func (iu *pChainInputUpdater) updateFromChain(notUpdated map[string][]shared.Input) error {
	fetchedOuts := make([]shared.Output, 0, 4*len(notUpdated))
	for txId := range notUpdated {
		tx, err := CallPChainGetTxApi(iu.client, txId)
		if err != nil {
			return err
		}

		txOuts := tx.Unsigned.Outputs()
		outs, err := shared.OutputsFromTxOuts(txId, txOuts, NewPChainTxOutput)
		if err != nil {
			return err
		}

		fetchedOuts = append(fetchedOuts, outs...)
	}
	return shared.UpdateInputsWithOutputs(notUpdated, fetchedOuts)
}
