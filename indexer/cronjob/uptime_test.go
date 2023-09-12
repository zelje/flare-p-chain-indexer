package cronjob

import (
	globalConfig "flare-indexer/config"
	"flare-indexer/database"
	"flare-indexer/indexer/config"
	"flare-indexer/indexer/context"
	"testing"
	"time"
)

func uptimeCronjobTestConfig() *config.Config {
	cfg := &config.Config{
		Chain: globalConfig.ChainConfig{
			ChainAddressHRP: "localflare",
			ChainID:         162,
		},
		UptimeCronjob: config.UptimeConfig{
			CronjobConfig: config.CronjobConfig{
				Enabled: true,
				Timeout: 60 * time.Second,
			},
		},
		DB: globalConfig.DBConfig{
			Username:   database.MysqlTestUser,
			Password:   database.MysqlTestPassword,
			Host:       database.MysqlTestHost,
			Port:       database.MysqlTestPort,
			Database:   "flare_indexer_indexer",
			LogQueries: false,
		},
	}
	return cfg

}

func createTestUptimeCronjob() (*uptimeCronjob, error) {
	ctx, err := context.BuildTestContext(uptimeCronjobTestConfig())
	if err != nil {
		return nil, err
	}
	return &uptimeCronjob{
		config: ctx.Config().UptimeCronjob,
		db:     ctx.DB(),
		client: testUptimeClient,
	}, nil
}

func TestUptime(t *testing.T) {
	cronjob, err := createTestUptimeCronjob()
	if err != nil {
		t.Fatal(err)
	}

	now := time.Unix(1675348249, 0)
	testUptimeClient.SetNow(now)

	for i := 0; i < 100; i++ {
		if err := cronjob.Call(); err != nil {
			t.Fatal(err)
		}
		testUptimeClient.Time.AdvanceNow(30 * time.Second)
	}

	uptimes, err := database.FetchUptimes(cronjob.db, []string{}, now, now.Add(31*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(uptimes) != 8 {
		t.Fatalf("expected 8 uptimes, got %d", len(uptimes))
	}
}
