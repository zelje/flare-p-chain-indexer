package main

import (
	"flare-indexer/logger"
	"flare-indexer/services/context"
	"flare-indexer/services/routes"
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

	router := mux.NewRouter()
	routes.AddValidatorRoutes(router, ctx)
	routes.AddQueryRoutes(router, ctx)

	address := ctx.Config().Services.Address
	srv := &http.Server{
		Handler: router,
		Addr:    address,
		// Good practice: enforce timeouts for servers you create -- config?
		// WriteTimeout: 15 * time.Second,
		// ReadTimeout:  15 * time.Second,
	}
	logger.Info("Starting server on %s", address)
	srv.ListenAndServe()
}
