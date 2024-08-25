// Package codes defines the error codes used by the application.
package codes

// Code represents an error code.
type Code int32

const (
	// OK is returned on success.
	OK Code = 0

	// Canceled indicates the operation was canceled.
	Canceled Code = 1

	// Unknown error.
	Unknown Code = 2

	// InvalidArgument indicates client specified an invalid argument.
	InvalidArgument Code = 3

	// DeadlineExceeded means operation expired before completion.
	DeadlineExceeded Code = 4

	// NotFound means some requested entity was not found.
	NotFound Code = 5

	// AlreadyExists means an attempt to create an entity failed because one already exists.
	AlreadyExists Code = 6

	// PermissionDenied indicates the caller does not have permission to execute the specified operation.
	PermissionDenied Code = 7

	// ResourceExhausted indicates some resource has been exhausted.
	ResourceExhausted Code = 8

	// FailedPrecondition indicates operation was rejected because the system is not in a state required for the operation's execution.
	FailedPrecondition Code = 9

	// Aborted indicates the operation was aborted.
	Aborted Code = 10

	// OutOfRange means operation was attempted past the valid range.
	OutOfRange Code = 11

	// Unimplemented indicates operation is not implemented or not supported/enabled.
	Unimplemented Code = 12

	// Internal errors.
	Internal Code = 13

	// Unavailable indicates the service is currently unavailable.
	Unavailable Code = 14

	// DataLoss indicates unrecoverable data loss or corruption.
	DataLoss Code = 15

	// Unauthenticated indicates the request does not have valid authentication credentials.
	Unauthenticated Code = 16
)

var codeToString = map[Code]string{
	OK:                 "OK",
	Canceled:           "Canceled",
	Unknown:            "Unknown",
	InvalidArgument:    "InvalidArgument",
	DeadlineExceeded:   "DeadlineExceeded",
	NotFound:           "NotFound",
	AlreadyExists:      "AlreadyExists",
	PermissionDenied:   "PermissionDenied",
	ResourceExhausted:  "ResourceExhausted",
	FailedPrecondition: "FailedPrecondition",
	Aborted:            "Aborted",
	OutOfRange:         "OutOfRange",
	Unimplemented:      "Unimplemented",
	Internal:           "Internal",
	Unavailable:        "Unavailable",
	DataLoss:           "DataLoss",
	Unauthenticated:    "Unauthenticated",
}

// String returns the string representation of the Code.
func (c Code) String() string {
	if s, ok := codeToString[c]; ok {
		return s
	}
	return "Unknown"
}

// IsValid returns true if the code is valid.
func (c Code) IsValid() bool {
	_, ok := codeToString[c]
	return ok
}
