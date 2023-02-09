package xchain

import (
	"flare-indexer/database"
	"flare-indexer/indexer/context"
	"flare-indexer/indexer/shared"
	"flare-indexer/utils"
	"flare-indexer/utils/chain"
	"fmt"

	"github.com/ava-labs/avalanchego/indexer"
	"github.com/ava-labs/avalanchego/vms/avm/txs"
	"github.com/ava-labs/avalanchego/wallet/chain/x"
	"gorm.io/gorm"
)

type xChainInputUpdater struct {
	shared.BaseInputUpdater

	db     *gorm.DB
	client indexer.Client
}

func newXChainInputUpdater(ctx context.IndexerContext, client indexer.Client) *xChainInputUpdater {
	ioUpdater := xChainInputUpdater{
		db:     ctx.DB(),
		client: client,
	}
	ioUpdater.InitCache(ctx.Config().Indexer.OutputsCacheSize)
	return &ioUpdater
}

func (iu *xChainInputUpdater) UpdateInputs(inputs map[string][]*database.TxInput) error {
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
func (iu *xChainInputUpdater) updateFromDB(notUpdated map[string][]*database.TxInput) error {
	outs, err := database.FetchXChainTxOutputs(iu.db, utils.Keys(notUpdated))
	if err != nil {
		return err
	}
	baseOuts := make([]*database.TxOutput, len(outs))
	for i, o := range outs {
		baseOuts[i] = &o.TxOutput
	}
	return shared.UpdateInputsWithOutputs(notUpdated, baseOuts)
}

// notUpdated is a map from *output* id to inputs referring this output
func (iu *xChainInputUpdater) updateFromChain(notUpdated map[string][]*database.TxInput) error {
	fetchedOuts := make([]*database.TxOutput, 0, 4*len(notUpdated))
	for txId := range notUpdated {
		container, err := chain.FetchContainerFromIndexer(iu.client, txId)
		if err != nil {
			return err
		}
		if container == nil {
			continue
		}

		tx, err := x.Parser.ParseGenesisTx(container.Bytes)
		if err != nil {
			return err
		}

		var outs []*database.TxOutput
		switch unsignedTx := tx.Unsigned.(type) {
		case *txs.BaseTx:
			outs, err = shared.TxOutputsFromBaseTx(txId, unsignedTx)
		case *txs.ImportTx:
			outs, err = shared.TxOutputsFromBaseTx(txId, &unsignedTx.BaseTx)
		default:
			return fmt.Errorf("transaction with id %s has unsupported type %T", container.ID.String(), unsignedTx)
		}
		if err != nil {
			return err
		}

		fetchedOuts = append(fetchedOuts, outs...)
	}
	return shared.UpdateInputsWithOutputs(notUpdated, fetchedOuts)
}
