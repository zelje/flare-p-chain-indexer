package main

import (
	"flare-indexer/indexer/context"
	"flare-indexer/indexer/migrations"
	"flare-indexer/indexer/runner"
	"fmt"
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
	runner.Start(ctx)
}
