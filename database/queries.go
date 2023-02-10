package database

import (
	"gorm.io/gorm"
)

func FetchState(db *gorm.DB, name string) (State, error) {
	var currentState State
	err := db.Where(&State{Name: name}).First(&currentState).Error
	return currentState, err
}

func FetchXChainTxOutputs(db *gorm.DB, ids []string) ([]XChainTxOutput, error) {
	var txs []XChainTxOutput
	err := db.Where("tx_id IN ?", ids).Find(&txs).Error
	return txs, err
}

func FetchPChainTxOutputs(db *gorm.DB, ids []string) ([]PChainTxOutput, error) {
	var txs []PChainTxOutput
	err := db.Where("tx_id IN ?", ids).Find(&txs).Error
	return txs, err
}

func FetchMigrations(db *gorm.DB) ([]Migration, error) {
	var migrations []Migration
	err := db.Order("version asc").Find(&migrations).Error
	return migrations, err
}

func CreateMigration(db *gorm.DB, m *Migration) error {
	return db.Create(m).Error
}

func UpdateMigration(db *gorm.DB, m *Migration) error {
	return db.Save(m).Error
}

func CreateState(db *gorm.DB, s *State) error {
	return db.Create(s).Error
}

func UpdateState(db *gorm.DB, s *State) error {
	return db.Save(s).Error
}

// TODO: assert the right types or merge with CreatePChainEntities
func CreateXChainEntities(db *gorm.DB, txs interface{}, ins interface{}, outs interface{}) error {
	var err error

	err = db.Create(&txs).Error
	if err != nil {
		return err
	}
	err = db.Create(&ins).Error
	if err != nil {
		return err
	}
	return db.Create(&outs).Error
}

// TODO: assert the right types or merge with CreateXChainEntities
func CreatePChainEntities(db *gorm.DB, txs interface{}, ins interface{}, outs interface{}) error {
	var err error

	err = db.Create(&txs).Error
	if err != nil {
		return err
	}
	err = db.Create(&ins).Error
	if err != nil {
		return err
	}
	return db.Create(&outs).Error
}
