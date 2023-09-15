package staking

import (
	"flare-indexer/config"
	"time"
)

type EpochInfo struct {
	Period time.Duration
	Start  time.Time
	First  int64
}

func NewEpochInfo(cfg *config.EpochConfig) EpochInfo {
	return EpochInfo{
		Period: cfg.Period,
		Start:  cfg.Start.Time,
		First:  cfg.First,
	}
}

func (e EpochInfo) GetStartTime(epoch int64) time.Time {
	return e.Start.Add(time.Duration(epoch) * e.Period)
}

func (e EpochInfo) GetEndTime(epoch int64) time.Time {
	return e.GetStartTime(epoch + 1)
}

func (e EpochInfo) GetTimeRange(epoch int64) (time.Time, time.Time) {
	return e.GetStartTime(epoch), e.GetEndTime(epoch)
}

func (e EpochInfo) GetEpochIndex(t time.Time) int64 {
	return int64(t.Sub(e.Start) / e.Period)
}
