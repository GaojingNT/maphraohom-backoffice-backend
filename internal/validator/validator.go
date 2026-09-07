package validator

import (
	"github.com/go-playground/validator"
	"maphraohom.app/maphraohom-backoffice/internal/exception"
)

func Validate(data interface{}) []exception.ParameterError {
	var (
		errParams []exception.ParameterError
		validate  = validator.New()
		err       = validate.Struct(data)
	)

	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element exception.ParameterError

			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()

			errParams = append(errParams, element)
		}
	}

	return errParams
}
