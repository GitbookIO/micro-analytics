package errors

var InternalError = RequestError{
    Code:       "InternalError",
    Message:    "We encountered an internal error. Please try again.",
    statusCode: 500,
}

var InvalidDatabaseName = RequestError{
    Code:       "InvalidDatabaseName",
    Message:    "Queried database doesn't exist.",
    statusCode: 400,
}

var InvalidJSON = RequestError{
    Code:       "InvalidJSON",
    Message:    "Invalid JSON in request body. Please check and retry.",
    statusCode: 400,
}