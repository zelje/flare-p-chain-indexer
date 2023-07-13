package main

import (
	"flare-indexer/logger"
	"flare-indexer/services/context"
	"flare-indexer/services/routes"
	"flare-indexer/services/utils"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	ctx, err := context.BuildContext()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	muxRouter := mux.NewRouter()
	router := utils.NewSwaggerRouter(muxRouter, "Flare P-Chain Indexer", "0.1.0")
	routes.AddTransferRoutes(router, ctx)
	routes.AddStakerRoutes(router, ctx)
	routes.AddTransactionRoutes(router, ctx)
	routes.AddQueryRoutes(router, ctx)

	router.Finalize()

	address := ctx.Config().Services.Address
	srv := &http.Server{
		Handler: muxRouter,
		Addr:    address,
		// Good practice: enforce timeouts for servers you create -- config?
		// WriteTimeout: 15 * time.Second,
		// ReadTimeout:  15 * time.Second,
	}
	logger.Info("Starting server on %s", address)
	srv.ListenAndServe()
}
