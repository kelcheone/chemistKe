// Package status implements error handling with rich status objects.
package status

import (
	"fmt"

	"github.com/kelcheone/chemistke/pkg/codes"
)

// Status represents an error with additional context.
type Status struct {
	code    codes.Code
	message string
	details []interface{}
}

// New creates a new Status.
func New(c codes.Code, msg string) *Status {
	return &Status{
		code:    c,
		message: msg,
	}
}

// Newf creates a new Status with formatted message.
func Newf(c codes.Code, format string, a ...interface{}) *Status {
	return New(c, fmt.Sprintf(format, a...))
}

// Error returns the string representation of the status.
func (s *Status) Error() string {
	return fmt.Sprintf("status code: %s, message: %s", s.code, s.message)
}

// Code returns the status code.
func (s *Status) Code() codes.Code {
	return s.code
}

// Message returns the status message.
func (s *Status) Message() string {
	return s.message
}

// WithDetails returns a new Status with the provided details appended.
func (s *Status) WithDetails(details ...interface{}) *Status {
	newStatus := &Status{
		code:    s.code,
		message: s.message,
		details: append([]interface{}{}, s.details...),
	}
	newStatus.details = append(newStatus.details, details...)
	return newStatus
}

// Details returns the status details.
func (s *Status) Details() []interface{} {
	return s.details
}

// FromError creates a Status from an error.
func FromError(err error) *Status {
	if err == nil {
		return nil
	}
	if se, ok := err.(*Status); ok {
		return se
	}
	return New(codes.Unknown, err.Error())
}

// Errorf creates a Status from the given code and format string.
func Errorf(c codes.Code, format string, a ...interface{}) error {
	return Newf(c, format, a...)
}

// Convert is a convenience function to convert errors to Status objects.
func Convert(err error) *Status {
	if err == nil {
		return nil
	}
	if se, ok := err.(*Status); ok {
		return se
	}
	return New(codes.Unknown, err.Error())
}

// OK returns a Status with OK code and no message.
func OK() *Status {
	return New(codes.OK, "")
}
