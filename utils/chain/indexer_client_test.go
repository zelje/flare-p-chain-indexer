package chain

import (
	"context"
	"testing"
)

func TestIndexerClient(t *testing.T) {
	client, err := PChainTestClient()
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.TODO()

	_, _, err = client.GetLastAccepted(ctx)
	if err != nil {
		t.Fatal(err)
	}

	cl, err := client.GetContainerRange(ctx, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(cl) != 10 {
		t.Fatalf("Expected 10 containers, got %d", len(cl))
	}

	c22, err := client.GetContainerByIndex(ctx, 22)
	if err != nil {
		t.Fatal(err)
	}

	cIdx, err := client.GetIndex(ctx, c22.ID)
	if err != nil {
		t.Fatal(err)
	}
	if cIdx != 22 {
		t.Fatalf("Expected index 22, got %d", cIdx)
	}

}
