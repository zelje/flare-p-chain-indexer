package routes

import (
	"flare-indexer/database"
	"flare-indexer/services/api"
	"flare-indexer/services/context"
	"flare-indexer/services/utils"
	"net/http"

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

func (rh *transactionRouteHandlers) getTransaction() utils.RouteHandler {
	handler := func(params map[string]string) (*api.ApiPChainTx, *utils.ErrorHandler) {
		txID := params["tx_id"]
		var resp *api.ApiPChainTx = nil
		err := database.DoInTransaction(rh.db, func(dbTx *gorm.DB) error {
			tx, inputs, outputs, err := database.FetchPChainTxFull(rh.db, txID)
			if err == nil {
				resp = api.NewApiPChainTx(tx, inputs, outputs)
			}
			return err
		})
		if err != nil {
			return nil, utils.InternalServerErrorHandler(err)
		}
		return resp, nil
	}
	return utils.NewParamRouteHandler(handler, http.MethodGet,
		map[string]string{"tx_id:[0-9a-zA-Z]+": "Transaction ID"},
		&api.ApiPChainTx{})
}

func AddTransactionRoutes(router utils.Router, ctx context.ServicesContext) {
	vr := newTransactionRouteHandlers(ctx)
	subrouter := router.WithPrefix("/transactions", "Transactions")
	subrouter.AddRoute("/get/{tx_id:[0-9a-zA-Z]+}", vr.getTransaction())
}
