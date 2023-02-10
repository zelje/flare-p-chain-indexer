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

func XChainTxInputFromTxInput(out *TxInput) *XChainTxInput {
	return &XChainTxInput{
		TxInput: *out,
	}
}

func PChainTxInputFromTxInput(out *TxInput) *PChainTxInput {
	return &PChainTxInput{
		TxInput: *out,
	}
}
