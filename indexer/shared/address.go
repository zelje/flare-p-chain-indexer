package shared

import (
	"flare-indexer/indexer/config"
	"fmt"

	"github.com/ava-labs/avalanchego/utils/formatting/address"
)

var (
	AddressHRP string
)

func init() {
	config.IndexerConfigCallback.AddCallback(func(config config.IndexerApplicationConfig) {
		AddressHRP = config.AddressHRP()
		if len(AddressHRP) == 0 {
			panic(fmt.Errorf("AddressHRP must be specified"))
		}
	})
}

func FormatAddressBytes(addr []byte) (string, error) {
	return address.FormatBech32(AddressHRP, addr)
}
