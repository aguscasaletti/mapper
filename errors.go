package mapper

import "fmt"

type FieldError struct {
	fieldName string
	context   string
	err       error
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("Invalid field: %v\n%v\n%v", e.fieldName, e.context, e.err.Error())
}

func NewFieldError(fieldName, context string, err error) *FieldError {
	return &FieldError{
		fieldName: fieldName,
		context:   context,
		err:       err,
	}
}

type ParameterError struct {
	parameterName string
	context       string
}

func (e *ParameterError) Error() string {
	return fmt.Sprintf("Invalid parameter: %v\n%v", e.parameterName, e.context)
}

func NewParamErrorNotNil(parameterName string) *ParameterError {
	return &ParameterError{
		parameterName: parameterName,
		context:       "cannot not be nil",
	}
}

var ErrTargetParamNotPointer = &ParameterError{
	parameterName: "target",
	context:       "must be a pointer",
}
