package database

func XChainTxOutputFromTxOutput(out *TxOutput) *XChainTxOutput {
	return &XChainTxOutput{
		TxOutput: *out,
	}
}

func PChainTxOutputFromTxOutput(out *TxOutput) *PChainTxOutput {
	return &PChainTxOutput{
		TxOutput: *out,
	}
}

func XChainTxInputFromTxInput(in *TxInput) *XChainTxInput {
	return &XChainTxInput{
		TxInput: *in,
	}
}

func PChainTxInputFromTxInput(in *TxInput) *PChainTxInput {
	return &PChainTxInput{
		TxInput: *in,
	}
}
