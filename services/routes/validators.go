package routes

import (
	"flare-indexer/database"
	"flare-indexer/services/api"
	"flare-indexer/services/context"
	"flare-indexer/services/utils"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type GetStakerRequest struct {
	NodeID    string
	Address   string
	StartTime time.Time
	EndTime   time.Time
}

type GetStakerResponse struct {
	TxIDs []string `json:"txIds"`
}

type validatorRouteHandlers struct {
	db *gorm.DB
}

func newValidatorRouteHandlers(ctx context.ServicesContext) *validatorRouteHandlers {
	return &validatorRouteHandlers{
		db: ctx.DB(),
	}
}

func (vr *validatorRouteHandlers) listValidators(w http.ResponseWriter, r *http.Request) {
	var request GetStakerRequest
	if !utils.DecodeBody(w, r, &request) {
		return
	}
	txIDs, err := database.FetchPChainValidators(vr.db, database.PChainAddValidatorTx, request.NodeID,
		request.Address, request.StartTime, request.EndTime, 0, 100)
	if utils.HandleInternalServerError(w, err) {
		return
	}
	utils.WriteApiResponseOk(w, GetStakerResponse{TxIDs: txIDs})
}

func (vr *validatorRouteHandlers) getValidator(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	txID := params["tx_id"]
	err := database.DoInTransaction(vr.db, func(dbTx *gorm.DB) error {
		tx, inputs, outputs, err := database.FetchPChainTxFull(vr.db, txID)
		if err == nil {
			utils.WriteApiResponseOk(w, api.NewApiPChainTx(tx, inputs, outputs))
		}
		return err
	})
	utils.HandleInternalServerError(w, err)
}

func AddValidatorRoutes(router *mux.Router, ctx context.ServicesContext) {
	vr := newValidatorRouteHandlers(ctx)
	subrouter := router.PathPrefix("/validators").Subrouter()

	subrouter.HandleFunc("/list", vr.listValidators).Methods(http.MethodPost)
	subrouter.HandleFunc("/get/{tx_id:[0-9a-zA-Z]+}", vr.getValidator).Methods(http.MethodGet)
}
