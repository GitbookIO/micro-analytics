package errors

import (
    "fmt"
)

type RequestError struct {
    statusCode  int
    Code        string  `json="code"`
    Message     string  `json="message"`
}

// Normalize RequestError statusCode
func (e *RequestError) StatusCode() int {
    if e.statusCode == 0 {
        return 200
    }
    return e.statusCode
}

// Implement Error interface
func (e *RequestError) Error() string {
    return e.Message
}

// Generate a RequestError from informations
func genError(code int, codeStr string, err string, a ...interface{}) *RequestError {
    return &RequestError{
        statusCode: code,
        Code:       codeStr,
        Message:    fmt.Sprintf(err, a...),
    }
}