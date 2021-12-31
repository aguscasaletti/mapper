package mapper

import (
	"errors"
	"fmt"
)

var (
	// ErrUnexpectedNil should not be nil
	ErrUnexpectedNil = errors.New("should not be nil")
	// ErrMustBePointer must be a pointer
	ErrMustBePointer = errors.New("must be a pointer")
)

// FieldError is produced at run-time while mapping values from one struct to another
type FieldError struct {
	fieldName string
	context   string
	err       error
}

func (e *FieldError) Error() string {
	return fmt.Sprintf("Invalid field: %v\n%v\n%v", e.fieldName, e.context, e.err.Error())
}

func newFieldError(fieldName, context string, err error) *FieldError {
	return &FieldError{
		fieldName: fieldName,
		context:   context,
		err:       err,
	}
}
