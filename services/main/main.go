package main

import (
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

	srv := &http.Server{
		Handler: router,
		Addr:    ctx.Config().Services.Address,
		// Good practice: enforce timeouts for servers you create -- config?
		// WriteTimeout: 15 * time.Second,
		// ReadTimeout:  15 * time.Second,
	}
	srv.ListenAndServe()
}
