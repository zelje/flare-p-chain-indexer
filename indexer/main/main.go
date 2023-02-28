package main

import (
	"flare-indexer/indexer/context"
	"flare-indexer/indexer/migrations"
	"flare-indexer/indexer/runner"
	"flare-indexer/indexer/shared"
	"flare-indexer/logger"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, err := context.BuildContext()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	err = migrations.Container.ExecuteAll(ctx)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	cancelChan := make(chan os.Signal, 1)
	signal.Notify(cancelChan, os.Interrupt, syscall.SIGTERM)

	// Prometheus metrics
	shared.InitMetricsServer(&ctx.Config().Metrics)

	runner.Start(ctx)

	<-cancelChan
	logger.Info("Stopped flare indexer")

}
