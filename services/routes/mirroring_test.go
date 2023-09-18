package routes

import (
	globalConfig "flare-indexer/config"
	"flare-indexer/database"
	"flare-indexer/services/api"
	"flare-indexer/services/config"
	serviceUtils "flare-indexer/services/utils"
	"flare-indexer/utils"
	"flare-indexer/utils/staking"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

var testMirroringData = map[string]database.PChainTxData{
	"2NuEmDJopBVunGZym7pcYjfuWTPaoWuHSnSvxiqdFdvDY7TGqQ": {
		PChainTx: database.PChainTx{
			Type:      database.PChainAddDelegatorTx,
			NodeID:    "NodeID-FQKTLuZHEsjCxPeFTFgsojsucmdyNDsz1",
			StartTime: pTime(2023, time.January, 1, 0, 0, 0, 0, time.UTC),
			EndTime:   pTime(2023, time.February, 1, 0, 0, 0, 0, time.UTC),
			TxID:      pString("2NuEmDJopBVunGZym7pcYjfuWTPaoWuHSnSvxiqdFdvDY7TGqQ"),
			Weight:    50000000000000,
		},
		InputAddress: "costwo1n5vvqn7g05sxzaes8xtvr5mx6m95q96jesrg5g",
	},
}

func TestMain(m *testing.M) {
	cfg := config.Config{
		Chain: globalConfig.ChainConfig{
			ChainAddressHRP: "costwo",
		},
	}
	globalConfig.GlobalConfigCallback.Call(cfg)
	m.Run()
}

func TestGetMirroringData(t *testing.T) {
	mh := newMirroringTestRouteHandlers(testMirroringData)

	r, err := http.NewRequest(http.MethodGet, "/tx_data/2NuEmDJopBVunGZym7pcYjfuWTPaoWuHSnSvxiqdFdvDY7TGqQ", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router := mux.NewRouter()
	router.HandleFunc("/tx_data/{tx_id}", mh.listMirroringTransactions().Handler)
	router.ServeHTTP(w, r)

	require.Equal(t, http.StatusOK, w.Result().StatusCode)

	var wResponse api.ApiResponseWrapper[GetMirroringResponse]
	serviceUtils.DecodeStruct(t, w.Result().Body, &wResponse)

	cupaloy.SnapshotT(t, wResponse)
}

func newMirroringTestRouteHandlers(txs map[string]database.PChainTxData) *mirroringRouteHandlers {
	return &mirroringRouteHandlers{
		db: newTestDB(txs),
		epochs: staking.NewEpochInfo(&globalConfig.EpochConfig{
			Period: 180 * time.Second,
			Start:  utils.Timestamp{Time: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)},
		}),
	}
}

type testDB struct {
	txs map[string]database.PChainTxData
}

func newTestDB(txs map[string]database.PChainTxData) mirrorDB {
	return &testDB{txs}
}

func (db testDB) FetchState(name string) (database.State, error) {
	return database.State{}, nil
}

func (db testDB) UpdateJobState(epoch int64) error {
	return nil
}

func (db testDB) GetPChainTxsForEpoch(start, end time.Time) ([]database.PChainTxData, error) {
	var txs []database.PChainTxData
	for _, tx := range db.txs {
		if !tx.StartTime.Before(start) && tx.StartTime.Before(end) {
			txs = append(txs, tx)
		}
	}
	return txs, nil
}

func (db testDB) GetPChainTx(txID string) (*database.PChainTx, error) {
	if tx, ok := db.txs[txID]; ok {
		return &tx.PChainTx, nil
	}
	return nil, gorm.ErrRecordNotFound
}

func pString(s string) *string { return &s }

func pTime(year int, month time.Month, day, hour, min, sec, nsec int, loc *time.Location) *time.Time {
	t := time.Date(year, month, day, hour, min, sec, nsec, loc)
	return &t
}
