// Package errors provides the standard error framework for CloudOS.
// It defines sentinel errors, typed error wrappers, and a stack-aware error type
// that carries structured context across subsystem boundaries.
package errors

import (
	"errors"
	"fmt"
)

// Code represents a machine-readable error classification.
type Code string

const (
	// General errors.
	CodeInternal       Code = "INTERNAL"
	CodeNotImplemented Code = "NOT_IMPLEMENTED"
	CodeTimeout        Code = "TIMEOUT"

	// Validation errors.
	CodeInvalidArgument Code = "INVALID_ARGUMENT"
	CodeInvalidState    Code = "INVALID_STATE"

	// Resource errors.
	CodeNotFound      Code = "NOT_FOUND"
	CodeAlreadyExists Code = "ALREADY_EXISTS"
	CodeConflict      Code = "CONFLICT"

	// Authorization errors.
	CodeUnauthenticated Code = "UNAUTHENTICATED"
	CodeForbidden       Code = "FORBIDDEN"

	// Resource exhaustion.
	CodeResourceExhausted Code = "RESOURCE_EXHAUSTED"
	CodeQuotaExceeded     Code = "QUOTA_EXCEEDED"

	// Dependency errors.
	CodeUnavailable Code = "UNAVAILABLE"
	CodeDependency  Code = "DEPENDENCY_FAILURE"
)

// CloudOSError is the standard error type for all CloudOS subsystems.
// It carries a machine-readable code, a human-readable message, and an optional
// wrapped cause.
type CloudOSError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"-"` // never serialise the cause stack
}

// Error implements the error interface.
func (e *CloudOSError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the wrapped cause error, if any.
func (e *CloudOSError) Unwrap() error {
	return e.Cause
}

// New creates a new CloudOSError with the given code and message.
func New(code Code, msg string) *CloudOSError {
	return &CloudOSError{Code: code, Message: msg}
}

// Newf creates a new CloudOSError with a formatted message.
func Newf(code Code, format string, args ...interface{}) *CloudOSError {
	return &CloudOSError{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Wrap wraps an existing error with additional context.
func Wrap(code Code, cause error, msg string) *CloudOSError {
	return &CloudOSError{Code: code, Message: msg, Cause: cause}
}

// Wrapf wraps an existing error with a formatted message.
func Wrapf(code Code, cause error, format string, args ...interface{}) *CloudOSError {
	return &CloudOSError{Code: code, Message: fmt.Sprintf(format, args...), Cause: cause}
}

// CodeOf extracts the error code from an error. Returns CodeInternal if the
// error is not a CloudOSError.
func CodeOf(err error) Code {
	if err == nil {
		return ""
	}
	var ce *CloudOSError
	if As(err, &ce) {
		return ce.Code
	}
	return CodeInternal
}

// MessageOf extracts the user-facing message from an error. Returns the error
// text if the error is not a CloudOSError.
func MessageOf(err error) string {
	if err == nil {
		return ""
	}
	var ce *CloudOSError
	if As(err, &ce) {
		return ce.Message
	}
	return err.Error()
}

// As is a convenience wrapper for errors.As.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// Is is used for sentinel comparison. For CloudOSError we compare by Code.
func Is(err error, code Code) bool {
	return CodeOf(err) == code
}

// Sentinel errors for common failure modes.
var (
	ErrNotFound      = New(CodeNotFound, "resource not found")
	ErrAlreadyExists = New(CodeAlreadyExists, "resource already exists")
	ErrInternal      = New(CodeInternal, "internal error")
	ErrInvalidInput  = New(CodeInvalidArgument, "invalid input")
	ErrUnauthenticated = New(CodeUnauthenticated, "not authenticated")
	ErrForbidden     = New(CodeForbidden, "forbidden")
	ErrNotImplemented = New(CodeNotImplemented, "not implemented")
	ErrTimeout       = New(CodeTimeout, "operation timed out")
)
