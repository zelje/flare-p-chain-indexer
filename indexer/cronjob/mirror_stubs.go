// Stubs for the mirror cronjob. These handle the direct interactions with DB
// and contracts. The actual logic is in mirror.go, which is unit-tested.
package cronjob

import (
	"flare-indexer/indexer/config"
	"flare-indexer/utils/contracts/mirroring"
	"flare-indexer/utils/contracts/voting"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pkg/errors"
)

type mirrorContractsCChain struct {
	mirroring *mirroring.Mirroring
	txOpts    *bind.TransactOpts
	voting    *voting.Voting
}

func initMirrorJobContracts(cfg *config.Config) (mirrorContracts, error) {
	if cfg.Mirror.MirroringContract == (common.Address{}) {
		return nil, errors.New("mirroring contract address not set")
	}

	if cfg.VotingCronjob.ContractAddress == (common.Address{}) {
		return nil, errors.New("voting contract address not set")
	}

	eth, err := ethclient.Dial(cfg.Chain.EthRPCURL)
	if err != nil {
		return nil, err
	}

	mirroringContract, err := mirroring.NewMirroring(cfg.Mirror.MirroringContract, eth)
	if err != nil {
		return nil, err
	}

	votingContract, err := voting.NewVoting(cfg.VotingCronjob.ContractAddress, eth)
	if err != nil {
		return nil, err
	}

	txOpts, err := TransactOptsFromPrivateKey(cfg.Chain.PrivateKey, cfg.Chain.ChainID)
	if err != nil {
		return nil, err
	}

	return &mirrorContractsCChain{
		mirroring: mirroringContract,
		txOpts:    txOpts,
		voting:    votingContract,
	}, nil
}

func (m mirrorContractsCChain) GetMerkleRoot(epoch int64) ([32]byte, error) {
	return m.voting.GetMerkleRoot(new(bind.CallOpts), big.NewInt(epoch))
}

func (m mirrorContractsCChain) MirrorStake(
	stakeData *mirroring.IPChainStakeMirrorVerifierPChainStake,
	merkleProof [][32]byte,
) error {
	_, err := m.mirroring.MirrorStake(m.txOpts, *stakeData, merkleProof)
	return err
}
