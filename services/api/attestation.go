package api

type VerificationStatus string

const (
	// Valid status
	VerificationStatusOK VerificationStatus = "OK"

	// Indeterminate statuses
	VerificationStatusDataAvailabilityIssue VerificationStatus = "DATA_AVAILABILITY_ISSUE"
	VerificationStatusNeedsMoreChecks       VerificationStatus = "NEEDS_MORE_CHECKS"
	VerificationStatusSystemFailure         VerificationStatus = "SYSTEM_FAILURE"
	VerificationStatusNonExistentBlock      VerificationStatus = "NON_EXISTENT_BLOCK"

	// Error statuses
	VerificationStatusNotConfirmed                    VerificationStatus = "NOT_CONFIRMED"
	VerificationStatusNotPayment                      VerificationStatus = "NOT_PAYMENT"
	VerificationStatusNotStandardPaymentReference     VerificationStatus = "NOT_STANDARD_PAYMENT_REFERENCE"
	VerificationStatusPaymentSummaryError             VerificationStatus = "PAYMENT_SUMMARY_ERROR"
	VerificationStatusReferencedTransactionExists     VerificationStatus = "REFERENCED_TRANSACTION_EXISTS"
	VerificationStatusZeroPaymentReferenceUnsupported VerificationStatus = "ZERO_PAYMENT_REFERENCE_UNSUPPORTED"
	VerificationStatusNonExistentTransaction          VerificationStatus = "NON_EXISTENT_TRANSACTION"
)

type SourceId int

const (
	SourceIdInvalid  SourceId = -1
	SourceIdBTC      SourceId = 0
	SourceIdLTC      SourceId = 1
	SourceIdDOGE     SourceId = 2
	SourceIdXRP      SourceId = 3
	SourceIdALGO     SourceId = 4
	SourceIdFLARE    SourceId = 14
	SourceIdSONGBIRD SourceId = 19
	SourceIdCOSTON   SourceId = 16
	SourceIdCOSTON2  SourceId = 114
)

type AttestationType int

const (
	AttestationTypePayment                       AttestationType = 1
	AttestationTypeBalanceDecreasingTransaction  AttestationType = 2
	AttestationTypeConfirmedBlockHeightExists    AttestationType = 3
	AttestationTypeReferencedPaymentNonexistence AttestationType = 4
	AttestationTypePChainStaking                 AttestationType = 5
)

// DTO object for posting attestation requests to verifier server
type APIAttestationRequest struct {

	// Attestation request in hex string representing byte sequence as submitted to State Connector smart contract.
	Request string `json:"request" validate:"required,hexadecimal"`
}

// DTO Object returned after attestation request verification.
// If status is 'OK' then fields Hash, Request and Response appear
// in the full response.
type APIVerification[R, T any] struct {
	// Hash of the attestation as included in Merkle tree.
	Hash string `json:"hash"`

	// Parsed attestation request.
	Request *R `json:"request"`

	// Attestation response.
	Response *T `json:"response"`

	// Verification status.
	Status VerificationStatus `json:"status"`
}

type DHPChainStaking struct {
	// Round id in which the attestation request was validated.
	StateConnectorRound uint64 `json:"stateConnectorRound"`

	// Merkle proof (a list of 32-byte hex hashes).
	MerkleProof []string `json:"merkleProof"`

	// Number of the transaction block on the underlying chain.
	BlockNumber uint64 `json:"blockNumber"`

	// Hash of the transaction on the underlying chain.
	TransactionHash string `json:"transactionHash"`

	// Type of the staking/delegation transaction: '0' for 'ADD_VALIDATOR_TX' and '1' for 'ADD_DELEGATOR_TX'.
	TransactionType byte `json:"transactionType"`

	// NodeID to which staking or delegation is done. For definitions, see https://github.com/ava-labs/avalanchego/blob/master/ids/node_id.go.
	NodeId string `json:"nodeId"`

	// Start time of the staking/delegation in seconds (Unix epoch).
	StartTime int64 `json:"startTime"`

	// End time of the staking/delegation in seconds (Unix epoch).
	EndTime int64 `json:"endTime"`

	// Staked or delegated amount in Gwei (nano FLR).
	Weight uint64 `json:"weight"`

	// Source address that triggered the staking or delegation transaction.
	// See https://support.avax.network/en/articles/4596397-what-is-an-address
	// for address definition for P-chain.
	SourceAddress string `json:"sourceAddress"`
}

type ARPChainStaking struct {

	// Attestation type id for this request, see 'AttestationType' enum.
	AttestationType AttestationType `json:"attestationType"`

	// The ID of the underlying chain, see 'SourceId' enum.
	SourceId SourceId `json:"sourceId"`

	// The hash of the expected attestation response appended by string 'Flare'. Used to verify consistency of the attestation response against the anticipated result, thus preventing wrong (forms of) attestations.
	MessageIntegrityCode string `json:"messageIntegrityCode"`

	// Transaction hash to search for.
	Id string `json:"id" validate:"required,tx-id"`

	// Block number of the transaction.
	BlockNumber uint64 `json:"blockNumber"`
}
