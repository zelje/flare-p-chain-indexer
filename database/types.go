package database

type TransactionType string

const (
	BaseTx   TransactionType = "BASE_TX"
	ImportTx TransactionType = "IMPORT_TX"
)

type MigrationStatus string

const (
	Pending   MigrationStatus = "PENDING"
	Completed MigrationStatus = "COMPLETED"
	Failed    MigrationStatus = "FAILED"
)
