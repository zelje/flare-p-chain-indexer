package routes

import (
	"flare-indexer/database"
	"flare-indexer/services/context"
	"flare-indexer/services/utils"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type GetTransferRequest struct {
	PaginatedRequest
	Address string `json:"address"`
}

type TxIDsResponse struct {
	TxIDs []string `json:"txIds"`
}

type transferRouteHandlers struct {
	db *gorm.DB
}

func newTransferRouteHandlers(ctx context.ServicesContext) *transferRouteHandlers {
	return &transferRouteHandlers{
		db: ctx.DB(),
	}
}

func (rh *transferRouteHandlers) listTransferTransactions(w http.ResponseWriter, r *http.Request, txType database.PChainTxType) {
	var request GetStakerRequest
	if !utils.DecodeBody(w, r, &request) {
		return
	}
	txIDs, err := database.FetchPChainTransferTransactions(rh.db, txType,
		request.Address, request.Offset, request.Limit)
	if utils.HandleInternalServerError(w, err) {
		return
	}
	utils.WriteApiResponseOk(w, GetStakerResponse{TxIDs: txIDs})
}

func AddTransferRoutes(router *mux.Router, ctx context.ServicesContext) {
	vr := newTransferRouteHandlers(ctx)
	importSubrouter := router.PathPrefix("/imports").Subrouter()

	importSubrouter.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		vr.listTransferTransactions(w, r, database.PChainImportTx)
	}).Methods(http.MethodPost)

	delegatorSubrouter := router.PathPrefix("/exports").Subrouter()

	delegatorSubrouter.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		vr.listTransferTransactions(w, r, database.PChainExportTx)
	}).Methods(http.MethodPost)
}
