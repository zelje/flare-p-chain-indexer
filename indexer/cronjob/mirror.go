package cronjob

import (
	"errors"
	"flare-indexer/database"
	"flare-indexer/indexer/config"
	"flare-indexer/indexer/context"
	"flare-indexer/utils/contracts/mirroring"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"gorm.io/gorm"
)

type mirrorCronJob struct {
	db                 *gorm.DB
	epochPeriodSeconds int
	mirroringContract  *mirroring.Mirroring
	epochTimeSeconds   int64
}

func NewMirrorCronJob(ctx context.IndexerContext) (Cronjob, error) {
	cfg := ctx.Config()
	mirroringContract, err := newMirroringContract(cfg)
	if err != nil {
		return nil, err
	}

	return &mirrorCronJob{
		db:                 ctx.DB(),
		epochPeriodSeconds: int(cfg.Mirror.EpochPeriod / time.Second),
		mirroringContract:  mirroringContract,
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
	epoch := c.getPreviousEpoch()
	if epoch < 0 {
		return errors.New("invalid epoch")
	}

	txs, err := c.getUnmirroredTxs(epoch)
	if err != nil {
		return err
	}

	if len(txs) == 0 {
		return nil
	}

	if err := c.mirrorTxs(txs); err != nil {
		return err
	}

	return c.markTxsAsMirrored(txs)
}

func (c *mirrorCronJob) getPreviousEpoch() int64 {
	currEpoch := (time.Now().Unix() - c.epochTimeSeconds) / int64(c.epochPeriodSeconds)
	return currEpoch - 1
}

func (c *mirrorCronJob) getUnmirroredTxs(epoch int64) ([]database.PChainTx, error) {
	startTimestamp := time.Duration(c.epochTimeSeconds+(epoch*int64(c.epochPeriodSeconds))) * time.Second
	endTimestamp := startTimestamp + (time.Duration(c.epochPeriodSeconds) * time.Second)

	var txs []database.PChainTx
	err := c.db.
		Where("mirrored = ?", false).
		Where("timestamp >= ?", startTimestamp).
		Where("timestamp < ?", endTimestamp).
		Find(&txs).
		Error
	if err != nil {
		return nil, err
	}

	return txs, nil
}

func (c *mirrorCronJob) mirrorTxs(txs []database.PChainTx) error {
	return errors.New("not implemented")
}

func (c *mirrorCronJob) markTxsAsMirrored(txs []database.PChainTx) error {
	for i := range txs {
		txs[i].Mirrored = true
	}

	return c.db.Save(&txs).Error
}
