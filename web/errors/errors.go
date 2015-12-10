package errors

var InsertFailed = RequestError{
	Code:       "InsertFailed",
	Message:    "Failed to insert your analytics. Please try again.",
	statusCode: 500,
}

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

var InvalidInterval = RequestError{
	Code:       "InvalidInterval",
	Message:    "Invalid interval format in request query. Please use specify a number in seconds and retry.",
	statusCode: 405,
}

var InvalidJSON = RequestError{
	Code:       "InvalidJSON",
	Message:    "Invalid JSON in request body. Please check and retry.",
	statusCode: 400,
}

var InvalidProperty = RequestError{
	Code:       "InvalidProperty",
	Message:    "Invalid request property. Please check and retry.",
	statusCode: 405,
}

var InvalidTimeFormat = RequestError{
	Code:       "InvalidTimeFormat",
	Message:    "Invalid time format in request query. Please use RFC3339 time and retry.",
	statusCode: 405,
}
