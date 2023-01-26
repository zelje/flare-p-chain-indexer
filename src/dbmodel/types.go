package dbmodel

type TransactionType string

const (
	BaseTx   TransactionType = "BASE_TX"
	ImportTx TransactionType = "IMPORT_TX"
)
