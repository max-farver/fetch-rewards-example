package common

import (
	"fmt"
	"github.com/pkg/errors"
)

type ErrorType uint

type HttpError struct {
	StatusCode int
	Message    string
}

const (
	// NoType error
	NoType ErrorType = iota
	// BadRequest error
	BadRequest
	// NotFound error
	NotFound
)

type rewardsError struct {
	errorType     ErrorType
	originalError error
	context       errorContext
}

type errorContext struct {
	Field   string
	Message string
}

// New creates a new rewardsError
func (errorType ErrorType) New(msg string) error {
	return rewardsError{errorType: errorType, originalError: errors.New(msg)}
}

// Newf creates a new rewardsError with formatted message
func (errorType ErrorType) Newf(msg string, args ...interface{}) error {
	return rewardsError{errorType: errorType, originalError: fmt.Errorf(msg, args...)}
}

// Wrap creates a new wrapped error
func (errorType ErrorType) Wrap(err error, msg string) error {
	return errorType.Wrapf(err, msg)
}

// Wrapf creates a new wrapped error with formatted message
func (errorType ErrorType) Wrapf(err error, msg string, args ...interface{}) error {
	return rewardsError{errorType: errorType, originalError: errors.Wrapf(err, msg, args...)}
}

// Error returns the message of a rewardsError
func (error rewardsError) Error() string {
	return error.originalError.Error()
}

// Wrap an error with a string
func Wrap(err error, msg string) error {
	return Wrapf(err, msg)
}

// Cause gives the original error
func Cause(err error) error {
	return errors.Cause(err)
}

// Wrapf an error with format string
func Wrapf(err error, msg string, args ...interface{}) error {
	wrappedError := errors.Wrapf(err, msg, args...)
	if customErr, ok := err.(rewardsError); ok {
		return rewardsError{
			errorType:     customErr.errorType,
			originalError: wrappedError,
			context:       customErr.context,
		}
	}

	return rewardsError{errorType: NoType, originalError: wrappedError}
}

// AddErrorContext adds a context to an error
func AddErrorContext(err error, field, message string) error {
	context := errorContext{Field: field, Message: message}
	if customErr, ok := err.(rewardsError); ok {
		return rewardsError{errorType: customErr.errorType, originalError: customErr.originalError, context: context}
	}

	return rewardsError{errorType: NoType, originalError: err, context: context}
}

// GetType returns the error type
func GetType(err error) ErrorType {
	if customErr, ok := err.(rewardsError); ok {
		return customErr.errorType
	}

	return NoType
}

// GetHttpError returns the error and corresponding status codes
func GetHttpError(err error) HttpError {
	switch GetType(err) {
	case NoType:
		return HttpError{StatusCode: 500, Message: err.Error()}
	case NotFound:
		return HttpError{StatusCode: 404, Message: err.Error()}
	case BadRequest:
		return HttpError{StatusCode: 400, Message: err.Error()}
	}
	return HttpError{StatusCode: 500, Message: err.Error()}
}
