package pchain

import (
	"flare-indexer/database"
	"flare-indexer/indexer/shared"
)

var (
	PChainDefaultInputOutputCreator = inputOutputCreator{outputType: database.PChainDefaultOutput}
	PChainStakerInputOutputCreator  = inputOutputCreator{outputType: database.PChainStakeOutput}
	PChainRewardOutputCreator       = inputOutputCreator{outputType: database.PChainRewardOutput}
)

type inputCreator struct{}

type inputOutputCreator struct {
	inputCreator
	outputType database.PChainOutputType
}

func (ioc inputCreator) CreateInput(in *database.TxInput) shared.Input {
	return &database.PChainTxInput{
		TxInput: *in,
	}
}

func (ioc inputOutputCreator) CreateOutput(out *database.TxOutput) shared.Output {
	return &database.PChainTxOutput{
		Type:     ioc.outputType,
		TxOutput: *out,
	}
}
