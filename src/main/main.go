package main

import (
	"flare-indexer/src/indexer"
	"flare-indexer/src/indexer/ctx"
	"flare-indexer/src/migrations"
	"fmt"
)

func main() {
	ctx, err := ctx.BuildContext()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	err = migrations.Container.ExecuteAll(ctx)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	indexer.Start(ctx)
}
