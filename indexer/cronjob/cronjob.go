package cronjob

import (
	"flare-indexer/logger"
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
