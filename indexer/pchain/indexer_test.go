package pchain

import (
	"flare-indexer/database"
	"flare-indexer/indexer/context"
	"testing"
)

func createPChainTestBlockIndexer(t *testing.T, batchSize int, startIndex uint64) *pChainBlockIndexer {
	ctx, err := context.BuildTestContext(pchainIndexerTestConfig(batchSize, startIndex))
	if err != nil {
		t.Fatal(err)
	}

	idxr := pChainBlockIndexer{}
	idxr.StateName = StateName
	idxr.IndexerName = "P-chain Blocks Test"
	idxr.Client = testClient
	idxr.DB = ctx.DB()
	idxr.Config = ctx.Config().PChainIndexer
	idxr.BatchIndexer = NewPChainBatchIndexer(ctx, idxr.Client, testRPCClient)

	return &idxr
}

func TestPChainBlockIndexerAll(t *testing.T) {
	idxr := createPChainTestBlockIndexer(t, 10, 0)

	// run one batch
	err := idxr.IndexBatch()
	if err != nil {
		t.Fatal(err)
	}

	txes, err := database.FetchTransactionsByBlockHeights(idxr.DB, []uint64{1, 2, 3, 4})
	if err != nil {
		t.Fatal(err)
	}

	if len(txes) != 4 {
		t.Fatalf("expected 4 txes, got %d", len(txes))
	}

	// run another batch
	err = idxr.IndexBatch()
	if err != nil {
		t.Fatal(err)
	}

	txes, err = database.FetchTransactionsByBlockHeights(idxr.DB, []uint64{16, 17, 18, 19, 20})
	if err != nil {
		t.Fatal(err)
	}

	if len(txes) != 5 {
		t.Fatalf("expected 5 txes, got %d", len(txes))
	}

}

// TestPChainBlockIndexerPartial tests that the indexer can handle a indexing
// from a non-zero start index
func TestPChainBlockIndexerPartial(t *testing.T) {
	idxr := createPChainTestBlockIndexer(t, 10, 20)

	// run one batch
	err := idxr.IndexBatch()
	if err != nil {
		t.Fatal(err)
	}

	txes, err := database.FetchTransactionsByBlockHeights(idxr.DB, []uint64{21, 22, 23, 24})
	if err != nil {
		t.Fatal(err)
	}

	if len(txes) != 4 {
		t.Fatalf("expected 4 txes, got %d", len(txes))
	}

	// run another batch
	err = idxr.IndexBatch()
	if err != nil {
		t.Fatal(err)
	}

	txes, err = database.FetchTransactionsByBlockHeights(idxr.DB, []uint64{26, 27, 28, 29, 30})
	if err != nil {
		t.Fatal(err)
	}

	if len(txes) != 5 {
		t.Fatalf("expected 5 txes, got %d", len(txes))
	}

}
