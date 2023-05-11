package chain

import (
	"testing"

	"github.com/ava-labs/avalanchego/ids"
)

func TestRPCClient(t *testing.T) {
	client := PChainRPCClient(t)
	id1, _ := ids.FromString("22ewQXuJw8PKQPiqJxwDezQszrNT2GbLyh4oCpCyVCSjAaDp2o")
	id2, _ := ids.FromString("oUpTu8TbYSWviCxV5mxuh2Wk9xSHRVrPXVKfPmFESPsRRdh2X")
	id3, _ := ids.FromString("2VhbseqzJLTZ1wxBWzWqvgshmAqx8LshT2p8HJP7P6zwz4iZTg")

	if id1.String() != "22ewQXuJw8PKQPiqJxwDezQszrNT2GbLyh4oCpCyVCSjAaDp2o" {
		t.Fatal("Wrong ID")
	}

	_, err := client.GetTx(id1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = client.GetTx(id2)
	if err == nil {
		t.Fatal("Expected error")
	}

	_, err = client.GetTx(id3)
	if err != nil {
		t.Fatal(err)
	}

	utxos1, err := client.GetRewardUTXOs(id1)
	if err != nil {
		t.Fatal(err)
	}
	if utxos1.NumFetched != 0 || len(utxos1.UTXOs) != 0 {
		t.Fatal("Expected 0 utxos")
	}

	utxos3, err := client.GetRewardUTXOs(id3)
	if err != nil {
		t.Fatal(err)
	}
	if utxos3.NumFetched != 2 || len(utxos3.UTXOs) != 2 {
		t.Fatal("Expected 2 utxos")
	}

}
