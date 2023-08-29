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

func ParseAddress(addr string) ([20]byte, error) {
	address20 := [20]byte{}
	hrp, address, err := address.ParseBech32(addr)
	if err != nil {
		return address20, err
	}
	if hrp != AddressHRP {
		return address20, fmt.Errorf("invalid address prefix: %s", hrp)
	}
	copy(address20[:], address)
	return address20, nil
}
