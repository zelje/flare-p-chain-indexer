package cronjob

import (
	"flare-indexer/indexer/config"
	"flare-indexer/logger"
	"flare-indexer/utils"
	"time"
)

type Cronjob interface {
	Name() string
	Enabled() bool
	Timeout() time.Duration
	Call() error
	OnStart() error
}

func RunCronjob(c Cronjob) {
	if !c.Enabled() {
		logger.Debug("%s cronjob disabled", c.Name())
		return
	}

	err := c.OnStart()
	if err != nil {
		logger.Error("%s cronjob on start error %v", c.Name(), err)
		return
	}

	logger.Debug("starting %s cronjob", c.Name())

	ticker := time.NewTicker(c.Timeout())
	for range ticker.C {
		err := c.Call()
		if err != nil {
			logger.Error("%s cronjob error %s", c.Name(), err.Error())
		}
	}
}

const (
	defaultEpochBatchSize int64 = 100
)

type epochCronjob struct {
	enabled   bool
	timeout   time.Duration
	epochs    epochInfo
	batchSize int64
}

type epochRange struct {
	start int64
	end   int64
}

func newEpochCronjob(cronjobCfg *config.CronjobConfig, epochCfg *config.EpochConfig) epochCronjob {
	return epochCronjob{
		enabled: cronjobCfg.Enabled,
		timeout: cronjobCfg.Timeout,
		epochs:  newEpochInfo(epochCfg),
	}
}

func (c *epochCronjob) Enabled() bool {
	return c.enabled
}

func (c *epochCronjob) Timeout() time.Duration {
	return c.timeout
}

// Get processing range (closed interval)
func (c *epochCronjob) getEpochRange(start int64, now time.Time) *epochRange {
	return c.getTrimmedEpochRange(start, c.epochs.getEpochIndex(now)-1)
}

// Get trimmed processing range (closed interval)
func (c *epochCronjob) getTrimmedEpochRange(start, end int64) *epochRange {
	start = utils.Max(start, c.epochs.first)
	batchSize := c.batchSize
	if batchSize <= 0 {
		batchSize = defaultEpochBatchSize
	}
	if end >= start+batchSize {
		end = batchSize + start - 1
	}
	return &epochRange{start, end}
}
