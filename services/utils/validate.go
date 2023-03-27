package utils

import (
	"flare-indexer/utils"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/go-playground/validator/v10"
)

var (
	validate *validator.Validate
)

func init() {
	validate = validator.New()
	validate.RegisterValidation("tx-id", ValidateTxID)
}

func ValidateTxID(fl validator.FieldLevel) bool {
	val := fl.Field().String()
	bytes, err := utils.DecodeHexString(val)
	if err != nil {
		return false
	}
	_, err = ids.ToID(bytes)
	return err == nil
}
