package routes

import (
	"errors"
	"flare-indexer/config"
	"flare-indexer/database"
	"flare-indexer/services/api"
	"flare-indexer/services/context"
	"flare-indexer/services/utils"
	globalUtils "flare-indexer/utils"
	"net/http"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type queryRouteHandlers struct {
	db  *gorm.DB
	cfg config.ChainConfig
}

func newQueryRouteHandlers(ctx context.ServicesContext) *queryRouteHandlers {
	return &queryRouteHandlers{
		db:  ctx.DB(),
		cfg: ctx.Config().ChainConfig(),
	}
}

func (qr *queryRouteHandlers) processAttestationRequest(w http.ResponseWriter, r *http.Request) {

}

// Request type: api.ARPChainStaking
//
// Response type: api.ApiResponseWrapper[api.APIVerification[api.ARPChainStaking, api.DHPChainStaking]]
func (qr *queryRouteHandlers) prepare(w http.ResponseWriter, r *http.Request) {
	var request api.ARPChainStaking
	if !utils.DecodeBody(w, r, &request) {
		return
	}
	if request.AttestationType != api.AttestationTypePChainStaking {
		utils.WriteApiResponseError(w, api.ApiResStatusInvalidRequest, "invalid attestation type",
			"attestation type must be pchain staking")
		return
	}
	if int(request.SourceId) != qr.cfg.ChainID {
		utils.WriteApiResponseError(w, api.ApiResStatusInvalidRequest, "invalid source chain id",
			"source chain id must be the same as the chain id of the indexer")
		return
	}

	// Ignore error, because it's already validated
	bytes, _ := globalUtils.DecodeHexString(request.Id)
	id, _ := ids.ToID(bytes)
	txID := id.String()

	tx, blockExists, err := database.FindPChainTxInBlockHeight(qr.db, txID, request.BlockNumber)
	if err != nil {
		utils.HandleInternalServerError(w, err)
		return
	}

	response := api.APIVerification[api.ARPChainStaking, api.DHPChainStaking]{}
	switch {
	case !blockExists:
		response.Status = api.VerificationStatusNonExistentBlock
	case tx == nil:
		response.Status = api.VerificationStatusNonExistentTransaction
	case tx.Type != database.PChainAddValidatorTx && tx.Type != database.PChainAddDelegatorTx:
		response.Status = api.VerificationStatusNonExistentTransaction
	default:
		var txType byte
		if tx.Type == database.PChainAddValidatorTx {
			txType = 0
		} else {
			txType = 1
		}

		// Ignore error, should be valid for add validator/delegator transactions
		nodeID, _ := globalUtils.NodeIDToHex(tx.NodeID)

		address, err := globalUtils.AddressToHex(tx.InputAddress)
		if err != nil {
			// Handle the case where the address is a node ID (genesis validator)
			address, err = globalUtils.IdToHex(tx.InputAddress)
			if err != nil {
				utils.HandleInternalServerError(w, errors.New("failed to convert address to hex"))
				return
			}
		}

		response.Status = api.VerificationStatusOK
		response.Request = &request
		response.Response = &api.DHPChainStaking{
			BlockNumber:     request.BlockNumber,
			TransactionHash: request.Id,
			TransactionType: txType,
			NodeId:          nodeID,
			StartTime:       tx.StartTime.Unix(),
			EndTime:         tx.EndTime.Unix(),
			Weight:          tx.Weight,
			SourceAddress:   address,
		}
	}
	utils.WriteApiResponseOk(w, response)
}

func AddQueryRoutes(router *mux.Router, ctx context.ServicesContext) {
	qr := newQueryRouteHandlers(ctx)
	subrouter := router.PathPrefix("/query").Subrouter()

	subrouter.HandleFunc("/", qr.processAttestationRequest).Methods(http.MethodPost)
	subrouter.HandleFunc("/prepare", qr.prepare).Methods(http.MethodPost)
}
