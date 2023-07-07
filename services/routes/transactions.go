package routes

import (
	"flare-indexer/database"
	"flare-indexer/services/api"
	"flare-indexer/services/context"
	"flare-indexer/services/utils"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type transactionRouteHandlers struct {
	db *gorm.DB
}

func newTransactionRouteHandlers(ctx context.ServicesContext) *transactionRouteHandlers {
	return &transactionRouteHandlers{
		db: ctx.DB(),
	}
}

func (rh *transactionRouteHandlers) getTransaction(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	txID := params["tx_id"]
	err := database.DoInTransaction(rh.db, func(dbTx *gorm.DB) error {
		tx, inputs, outputs, err := database.FetchPChainTxFull(rh.db, txID)
		if err == nil {
			utils.WriteApiResponseOk(w, api.NewApiPChainTx(tx, inputs, outputs))
		}
		return err
	})
	utils.HandleInternalServerError(w, err)
}

func AddTransactionRoutes(router *mux.Router, ctx context.ServicesContext) {
	vr := newTransactionRouteHandlers(ctx)
	subrouter := router.PathPrefix("/transactions").Subrouter()
	subrouter.HandleFunc("/get/{tx_id:[0-9a-zA-Z]+}", vr.getTransaction).Methods(http.MethodGet)
}
