// Package errors provides error handling utilities for RPC services.
// It includes error wrapping, error codes, and gRPC status mapping.
package errors

import (
	"errors"
	"fmt"

	"github.com/cloudwego/kitex/pkg/kerrors"
)

// Common error codes for RPC services.
const (
	// CodeOK indicates success.
	CodeOK = 0
	// CodeUnknown indicates an unknown error.
	CodeUnknown = 1
	// CodeInvalidArgument indicates invalid input.
	CodeInvalidArgument = 2
	// CodeNotFound indicates resource not found.
	CodeNotFound = 3
	// CodeAlreadyExists indicates resource already exists.
	CodeAlreadyExists = 4
	// CodePermissionDenied indicates permission denied.
	CodePermissionDenied = 5
	// CodeUnauthenticated indicates unauthenticated request.
	CodeUnauthenticated = 6
	// CodeResourceExhausted indicates resource exhausted (e.g., rate limit).
	CodeResourceExhausted = 7
	// CodeFailedPrecondition indicates a failed precondition.
	CodeFailedPrecondition = 8
	// CodeAborted indicates the operation was aborted.
	CodeAborted = 9
	// CodeOutOfRange indicates out of range error.
	CodeOutOfRange = 10
	// CodeUnimplemented indicates the operation is not implemented.
	CodeUnimplemented = 11
	// CodeInternal indicates an internal server error.
	CodeInternal = 12
	// CodeUnavailable indicates the service is unavailable.
	CodeUnavailable = 13
	// CodeDeadlineExceeded indicates the deadline was exceeded.
	CodeDeadlineExceeded = 14
	// CodeCancelled indicates the operation was cancelled.
	CodeCancelled = 15
)

// Error represents an RPC error with code and message.
type Error struct {
	Code    int32
	Message string
	cause   error
}

// New creates a new Error with the given code and message.
func New(code int32, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// Newf creates a new Error with the given code and formatted message.
func Newf(code int32, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

// Wrap wraps an existing error with an RPC error code.
func Wrap(err error, code int32, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		cause:   err,
	}
}

// Wrapf wraps an existing error with an RPC error code and formatted message.
func Wrapf(err error, code int32, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		cause:   err,
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("code=%d, message=%s, cause=%v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("code=%d, message=%s", e.Code, e.Message)
}

// Unwrap returns the underlying cause.
func (e *Error) Unwrap() error {
	return e.cause
}

// Is checks if the error matches the target.
func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return e.Code == t.Code
	}
	return false
}

// Common error constructors

// InvalidArgument returns an error indicating invalid input.
func InvalidArgument(message string) *Error {
	return New(CodeInvalidArgument, message)
}

// InvalidArgumentf returns an error indicating invalid input with formatted message.
func InvalidArgumentf(format string, args ...interface{}) *Error {
	return Newf(CodeInvalidArgument, format, args...)
}

// NotFound returns an error indicating resource not found.
func NotFound(message string) *Error {
	return New(CodeNotFound, message)
}

// NotFoundf returns an error indicating resource not found with formatted message.
func NotFoundf(format string, args ...interface{}) *Error {
	return Newf(CodeNotFound, format, args...)
}

// AlreadyExists returns an error indicating resource already exists.
func AlreadyExists(message string) *Error {
	return New(CodeAlreadyExists, message)
}

// PermissionDenied returns an error indicating permission denied.
func PermissionDenied(message string) *Error {
	return New(CodePermissionDenied, message)
}

// Unauthenticated returns an error indicating unauthenticated request.
func Unauthenticated(message string) *Error {
	return New(CodeUnauthenticated, message)
}

// Internal returns an error indicating internal server error.
func Internal(message string) *Error {
	return New(CodeInternal, message)
}

// Internalf returns an error indicating internal server error with formatted message.
func Internalf(format string, args ...interface{}) *Error {
	return Newf(CodeInternal, format, args...)
}

// Unavailable returns an error indicating service unavailable.
func Unavailable(message string) *Error {
	return New(CodeUnavailable, message)
}

// DeadlineExceeded returns an error indicating deadline exceeded.
func DeadlineExceeded(message string) *Error {
	return New(CodeDeadlineExceeded, message)
}

// FromError extracts an Error from an error.
// If the error is not an Error, it returns nil.
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return nil
}

// Code extracts the error code from an error.
// Returns CodeUnknown if the error is not an Error.
func Code(err error) int32 {
	if err == nil {
		return CodeOK
	}
	if e := FromError(err); e != nil {
		return e.Code
	}
	return CodeUnknown
}

// IsCode checks if the error has the specified code.
func IsCode(err error, code int32) bool {
	return Code(err) == code
}

// IsNotFound checks if the error indicates resource not found.
func IsNotFound(err error) bool {
	return IsCode(err, CodeNotFound)
}

// IsInvalidArgument checks if the error indicates invalid input.
func IsInvalidArgument(err error) bool {
	return IsCode(err, CodeInvalidArgument)
}

// IsUnauthenticated checks if the error indicates unauthenticated request.
func IsUnauthenticated(err error) bool {
	return IsCode(err, CodeUnauthenticated)
}

// IsPermissionDenied checks if the error indicates permission denied.
func IsPermissionDenied(err error) bool {
	return IsCode(err, CodePermissionDenied)
}

// IsInternal checks if the error indicates internal server error.
func IsInternal(err error) bool {
	return IsCode(err, CodeInternal)
}

// ToKitexError converts an Error to a Kitex error.
func ToKitexError(err *Error) error {
	if err == nil {
		return nil
	}
	return kerrors.NewBizStatusError(err.Code, err.Message)
}

// FromKitexError extracts an Error from a Kitex error.
func FromKitexError(err error) *Error {
	if err == nil {
		return nil
	}
	var bizErr kerrors.BizStatusErrorIface
	if errors.As(err, &bizErr) {
		return New(bizErr.BizStatusCode(), bizErr.BizMessage())
	}
	return nil
}