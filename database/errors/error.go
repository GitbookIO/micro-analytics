package errors

import (
	"fmt"
)

type DriverError struct {
	Code    int
	Message string
}

// Implement Error interface
func (e *DriverError) Error() string {
	return fmt.Sprintf("%d: %s", e.Code, e.Message)
}

// Produce RequestError
func Errorf(code int, err string, a ...interface{}) *DriverError {
	return &DriverError{
		Code:    code,
		Message: fmt.Sprintf(err, a...),
	}
}
