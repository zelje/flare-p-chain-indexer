package merkle

import (
	"errors"
	"sort"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

var (
	ErrEmptyTree    = errors.New("empty tree")
	ErrInvalidIndex = errors.New("invalid index")
)

// MerkleTree implementation with helper functions.
type MerkleTree struct {
	tree []common.Hash
}

// New creates a new Merkle tree from the given hash values as bytes.
func New(values []common.Hash) MerkleTree {
	return MerkleTree{tree: values}
}

// NewFromHex creates a new Merkle tree from the given hex values.
func NewFromHex(hexValues []string) MerkleTree {
	values := make([]common.Hash, len(hexValues))

	for i, hexValue := range hexValues {
		values[i] = common.HexToHash(hexValue)
	}

	return New(values)
}

// Given an array of leaf hashes, builds the Merkle tree.
func Build(values []common.Hash, initialHash bool) MerkleTree {
	hexValues := make([]string, len(values))

	for i, value := range values {
		hexValues[i] = value.Hex()
	}

	return BuildFromHex(hexValues, initialHash)
}

// Given an array of hex-encoded leaf hashes, builds the Merkle tree.
func BuildFromHex(hexValues []string, initialHash bool) MerkleTree {
	sort.Strings(hexValues)

	var hashes []common.Hash
	for i := range hexValues {
		if i == 0 || hexValues[i] != hexValues[i-1] {
			hashes = append(hashes, common.HexToHash(hexValues[i]))
		}
	}

	if initialHash {
		hashes = mapSingleHash(hashes)
	}

	n := len(hashes)
	tree := make([]common.Hash, n-1, (2*n)-1)
	tree = append(tree, hashes...)

	for i := n - 2; i >= 0; i-- {
		tree[i] = SortedHashPair(tree[2*i+1], tree[2*i+2])
	}

	return New(tree)
}

func mapSingleHash(hashes []common.Hash) []common.Hash {
	output := make([]common.Hash, len(hashes))

	for i := range hashes {
		output[i] = crypto.Keccak256Hash(hashes[i].Bytes())
	}

	return output
}

// SortedHashPair returns a sorted hash of two hashes.
func SortedHashPair(x, y common.Hash) common.Hash {
	if x.Hex() <= y.Hex() {
		return crypto.Keccak256Hash(x.Bytes(), y.Bytes())
	}

	return crypto.Keccak256Hash(y.Bytes(), x.Bytes())
}

// Root returns the Merkle root of the tree.
func (t MerkleTree) Root() (common.Hash, error) {
	if len(t.tree) == 0 {
		return common.Hash{}, ErrEmptyTree
	}

	return t.tree[0], nil
}

// Tree returns the a slice representing the full tree.
func (t MerkleTree) Tree() []common.Hash {
	return t.tree
}

// HashCount returns the number of leaves in the tree.
func (t MerkleTree) HashCount() int {
	if len(t.tree) == 0 {
		return 0
	}

	return (len(t.tree) + 1) / 2
}

// SortedHashes returns all leaves in a slice.
func (t MerkleTree) SortedHashes() []common.Hash {
	numLeaves := t.HashCount()
	if numLeaves == 0 {
		return nil
	}

	return t.tree[numLeaves-1:]
}

// GetHash returns the hash of the `i`th leaf.
func (t MerkleTree) GetHash(i int) (common.Hash, error) {
	numLeaves := t.HashCount()
	if numLeaves == 0 || i < 0 || i >= numLeaves {
		return common.Hash{}, ErrInvalidIndex
	}

	pos := len(t.tree) - numLeaves + i
	return t.tree[pos], nil
}

// GetProof returns the Merkle proof for the `i`th leaf.
func (t MerkleTree) GetProof(i int) ([]common.Hash, error) {
	numLeaves := t.HashCount()
	if numLeaves == 0 || i < 0 || i >= numLeaves {
		return nil, ErrInvalidIndex
	}

	var proof []common.Hash

	for pos := len(t.tree) - numLeaves + i; pos > 0; pos = parent(pos) {
		sibling := pos + ((2 * (pos % 2)) - 1)
		proof = append(proof, t.tree[sibling])
	}

	return proof, nil
}

// parent returns the index of the parent node.
func parent(i int) int {
	return (i - 1) / 2
}

// VerifyProof verifies a Merkle proof for a given leaf.
func VerifyProof(leaf common.Hash, proof []common.Hash, root common.Hash) bool {
	hash := leaf
	for _, pair := range proof {
		hash = SortedHashPair(pair, hash)
	}

	return hash == root
}
