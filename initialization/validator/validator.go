package validator

import (
	"planA/initialization/golabl"

	"github.com/go-playground/validator/v10"
)

func Init() {
	golabl.Validator = validator.New()
}
