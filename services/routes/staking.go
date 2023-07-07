package routes

import (
	"flare-indexer/database"
	"flare-indexer/services/context"
	"flare-indexer/services/utils"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type GetStakerRequest struct {
	PaginatedRequest
	NodeID  string    `json:"nodeId"`
	Address string    `json:"address"`
	Time    time.Time `json:"time"`
}

type GetStakerResponse struct {
	TxIDs []string `json:"txIds"`
}

type stakerRouteHandlers struct {
	db *gorm.DB
}

func newStakerRouteHandlers(ctx context.ServicesContext) *stakerRouteHandlers {
	return &stakerRouteHandlers{
		db: ctx.DB(),
	}
}

func (rh *stakerRouteHandlers) listStakingTransactions(w http.ResponseWriter, r *http.Request, txType database.PChainTxType) {
	var request GetStakerRequest
	if !utils.DecodeBody(w, r, &request) {
		return
	}
	txIDs, err := database.FetchPChainStakingTransactions(rh.db, txType, request.NodeID,
		request.Address, request.Time, request.Offset, request.Limit)
	if utils.HandleInternalServerError(w, err) {
		return
	}
	utils.WriteApiResponseOk(w, GetStakerResponse{TxIDs: txIDs})
}

func AddStakerRoutes(router *mux.Router, ctx context.ServicesContext) {
	vr := newStakerRouteHandlers(ctx)
	validatorSubrouter := router.PathPrefix("/validators").Subrouter()

	validatorSubrouter.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		vr.listStakingTransactions(w, r, database.PChainAddValidatorTx)
	}).Methods(http.MethodPost)

	delegatorSubrouter := router.PathPrefix("/delegators").Subrouter()

	delegatorSubrouter.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		vr.listStakingTransactions(w, r, database.PChainAddDelegatorTx)
	}).Methods(http.MethodPost)
}
