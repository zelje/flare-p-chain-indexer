package utils

import (
	"flare-indexer/config"
	"fmt"

	"github.com/ava-labs/avalanchego/utils/formatting/address"
)

var (
	AddressHRP string
)

func init() {
	config.GlobalConfigCallback.AddCallback(func(config config.GlobalConfig) {
		AddressHRP = config.ChainConfig().ChainAddressHRP
		if len(AddressHRP) == 0 {
			panic(fmt.Errorf("AddressHRP must be specified"))
		}
	})
}

func FormatAddressBytes(addr []byte) (string, error) {
	return address.FormatBech32(AddressHRP, addr)
}
