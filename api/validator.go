package api

import (
	"github.com/go-playground/validator/v10"
	"github.com/kjasn/simple-bank/utils"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	if value, ok := fl.Field().Interface().(string); ok {
		// check currency is supported
		if utils.IsSupportedCurrency(value) {
			return true
		}
	}


	return false 
}
