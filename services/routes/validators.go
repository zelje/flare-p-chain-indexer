package routes

import (
	"encoding/json"
	"flare-indexer/database"
	"flare-indexer/services/context"
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
	TxIDs []string
}

type validatorRouteHandlers struct {
	db *gorm.DB
}

func newValidatorRouteHandlers(ctx context.ServicesContext) *validatorRouteHandlers {
	return &validatorRouteHandlers{
		db: ctx.DB(),
	}
}

func (vr *validatorRouteHandlers) getValidators(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var request GetStakerRequest
	err := decoder.Decode(&request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	txIDs, err := database.FetchPChainValidators(vr.db, request.NodeID, request.Address, request.StartTime,
		request.EndTime, 0, 100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := GetStakerResponse{TxIDs: txIDs}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&response)
}

func AddValidatorRoutes(router *mux.Router, ctx context.ServicesContext) {
	vr := newValidatorRouteHandlers(ctx)

	router.HandleFunc("/validators", vr.getValidators).Methods(http.MethodPost)
}
