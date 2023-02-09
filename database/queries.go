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

func CreateXChainEntities(db *gorm.DB, txs []*XChainTx, ins []*XChainTxInput, outs []*XChainTxOutput) error {
	var err error

	err = db.Create(&txs).Error
	if err != nil {
		return err
	}
	err = db.Create(&ins).Error
	if err != nil {
		return err
	}
	err = db.Create(&outs).Error
	if err != nil {
		return err
	}
	return nil
}

func CreatePChainEntities(db *gorm.DB, txs []*PChainTx, ins []*PChainTxInput, outs []*PChainTxOutput, stakeOuts []*PChainStakeOutput) error {
	var err error

	err = db.Create(&txs).Error
	if err != nil {
		return err
	}
	err = db.Create(&ins).Error
	if err != nil {
		return err
	}
	err = db.Create(&outs).Error
	if err != nil {
		return err
	}
	return db.Create(&stakeOuts).Error
}
