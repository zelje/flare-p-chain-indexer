package chain

import (
	"testing"
	"time"

	"golang.org/x/exp/slices"
)

func TestUptimeClient(t *testing.T) {
	client, err := UptimeTestClient()
	if err != nil {
		t.Fatal(err)
	}

	// List all validators at 2023-02-02 14:00:00 UTC
	client.SetNow(time.Date(2023, time.February, 2, 14, 0, 0, 0, time.UTC))
	validators, _, err := client.GetValidatorStatus()
	if err != nil {
		t.Fatal(err)
	}
	if len(validators) != 4 {
		t.Fatalf("expected 4 validators, got %d", len(validators))
	}

	// Check if "NodeID-MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ" is not connected at 1676629054
	client.SetNowUnix(1676629054)
	validators, _, err = client.GetValidatorStatus()
	if err != nil {
		t.Fatal(err)
	}
	if len(validators) != 3 {
		t.Fatalf("expected 3 validators, got %d", len(validators))
	}

	index := slices.IndexFunc(validators, func(v ValidatorStatus) bool {
		return v.NodeID == "NodeID-MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ"
	})
	if index < 0 {
		t.Fatalf("expected NodeID-MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ to be in the list")
	}
	if validators[index].Connected {
		t.Fatalf("expected NodeID-MFrZFVCXPv5iCn6M9K6XduxGTYp891xXZ to be disconnected")
	}
}
