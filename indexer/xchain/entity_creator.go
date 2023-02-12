package xchain

import (
	"flare-indexer/database"
	"flare-indexer/indexer/shared"
)

var XChainInputOutputCreator = inputOutputCreator{}

type inputOutputCreator struct{}

func (ioc inputOutputCreator) CreateInput(in *database.TxInput) shared.Input {
	return &database.XChainTxInput{
		TxInput: *in,
	}
}

func (ioc inputOutputCreator) CreateOutput(out *database.TxOutput) shared.Output {
	return &database.XChainTxOutput{
		TxOutput: *out,
	}
}
