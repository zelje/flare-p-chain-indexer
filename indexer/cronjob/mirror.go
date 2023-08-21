package cronjob

import (
	"errors"
	"flare-indexer/indexer/config"
	"flare-indexer/indexer/context"
	"flare-indexer/utils/contracts/mirroring"

	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

type mirrorCronJob struct {
	db                 *gorm.DB
	mirroringContract  *mirroring.Mirroring
	epochPeriodSeconds int
}

func NewMirrorCronJob(ctx context.IndexerContext) (Cronjob, error) {
	mirroringContract, err := newMirroringContract(ctx.Config())
	if err != nil {
		return nil, err
	}

	return &mirrorCronJob{
		db:                ctx.DB(),
		mirroringContract: mirroringContract,
	}, nil
}

func newMirroringContract(cfg *config.Config) (*mirroring.Mirroring, error) {
	eth, err := ethclient.Dial(cfg.Chain.EthRPCURL)
	if err != nil {
		return nil, err
	}

	return mirroring.NewMirroring(cfg.Mirror.MirroringContract, eth)
}

func (c *mirrorCronJob) Name() string {
	return "mirror"
}

func (c *mirrorCronJob) Enabled() bool {
	return true
}

func (c *mirrorCronJob) TimeoutSeconds() int {
	return c.epochPeriodSeconds
}

func (c *mirrorCronJob) Call() error {
	return errors.New("not implemented")
}
