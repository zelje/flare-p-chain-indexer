package chain

import "github.com/ava-labs/avalanchego/utils/formatting/address"

func FormatAddressBytes(addr []byte) (string, error) {
	return address.FormatBech32("localflare", addr)
}
