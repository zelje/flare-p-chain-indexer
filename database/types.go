package database

// X-chain types

type XChainTxType string

const (
	XChainBaseTx   XChainTxType = "BASE_TX"
	XChainImportTx XChainTxType = "IMPORT_TX"
)

// P-chain types

type PChainTxType string

const (
	PChainRewardValidatorTx PChainTxType = "REWARD_TX"
	PChainAddDelegatorTx    PChainTxType = "ADD_DELEGATOR_TX"
	PChainAddValidatorTx    PChainTxType = "ADD_VALIDATOR_TX"
	PChainImportTx          PChainTxType = "IMPORT_TX"
	PChainExportTx          PChainTxType = "EXPORT_TX"
)

// Misc other types

type MigrationStatus string

const (
	MigrationPending   MigrationStatus = "PENDING"
	MigrationCompleted MigrationStatus = "COMPLETED"
	MigrationFailed    MigrationStatus = "FAILED"
)
