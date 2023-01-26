package main

import (
	"flare-indexer/src/indexer"
	"fmt"
)

func main() {
	ctx, err := indexer.BuildContext()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	indexer.Start(ctx)
}
