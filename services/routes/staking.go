package routes

import (
	"flare-indexer/database"
	"flare-indexer/services/context"
	"flare-indexer/services/utils"
	"net/http"
	"time"

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

func (rh *stakerRouteHandlers) listStakingTransactions(txType database.PChainTxType) utils.RouteHandler {
	handler := func(request GetStakerRequest) (GetStakerResponse, *utils.ErrorHandler) {
		txIDs, err := database.FetchPChainStakingTransactions(rh.db, txType, request.NodeID,
			request.Address, request.Time, request.Offset, request.Limit)
		if err != nil {
			return GetStakerResponse{}, utils.InternalServerErrorHandler(err)
		}
		return GetStakerResponse{TxIDs: txIDs}, nil
	}
	return utils.NewRouteHandler(handler, http.MethodPost, GetStakerRequest{}, GetStakerResponse{})
}

func AddStakerRoutes(router utils.Router, ctx context.ServicesContext) {
	vr := newStakerRouteHandlers(ctx)

	validatorSubrouter := router.WithPrefix("/validators", "Staking")
	validatorSubrouter.AddRoute("/transactions", vr.listStakingTransactions(database.PChainAddValidatorTx))

	delegatorSubrouter := router.WithPrefix("/delegators", "Staking")
	delegatorSubrouter.AddRoute("/transactions", vr.listStakingTransactions(database.PChainAddDelegatorTx))
}
