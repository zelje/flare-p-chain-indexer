package utils

import (
	"encoding/hex"
	"errors"
	"strings"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/avalanchego/utils/formatting/address"
)

const (
	hexPrefix             = "0x"
	addressChainSeparator = "-"
)

// DecodeHexString decodes a string that is prefixed with "0x" into a byte slice
func DecodeHexString(s string) ([]byte, error) {
	if !strings.HasPrefix(s, hexPrefix) {
		return nil, errors.New("string does not have hex prefix")
	}
	return hex.DecodeString(s[len(hexPrefix):])
}

// Convert node id string to 20 byte hex string
func NodeIDToHex(nodeID string) (string, error) {
	id, err := ids.NodeIDFromString(nodeID)
	if err != nil {
		return "", err
	}
	return hexPrefix + hex.EncodeToString(id.Bytes()), nil
}

// Convert address string to 20 byte hex string
func AddressToHex(addrStr string) (string, error) {
	if !strings.Contains(addrStr, addressChainSeparator) {
		addrStr = addressChainSeparator + addrStr
	}
	id, err := address.ParseToID(addrStr)
	if err != nil {
		return "", err
	}
	return hexPrefix + hex.EncodeToString(id.Bytes()), nil
}

// Convert id string to 20 byte hex string
func IdToHex(idStr string) (string, error) {
	id, err := ids.FromString(idStr)
	if err != nil {
		return "", err
	}
	return hexPrefix + hex.EncodeToString(id[:]), nil
}
