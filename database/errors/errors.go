package errors

var InternalError = DriverError{
	Code:    1,
	Message: "Internal error",
}

var InvalidDatabaseName = DriverError{
	Code:    2,
	Message: "Invalid database name",
}

var InsertFailed = DriverError{
	Code:    3,
	Message: "Failed to insert into DB",
}
