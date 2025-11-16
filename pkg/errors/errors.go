package errors

import (
	"fmt"
)

// ErrorType represents the category of error
type ErrorType string

const (
	ErrorTypeConfig       ErrorType = "CONFIG"
	ErrorTypeUpstream     ErrorType = "UPSTREAM"
	ErrorTypeProxy        ErrorType = "PROXY"
	ErrorTypeValidation   ErrorType = "VALIDATION"
	ErrorTypeRateLimit    ErrorType = "RATE_LIMIT"
	ErrorTypeInternal     ErrorType = "INTERNAL"
	ErrorTypeCache        ErrorType = "CACHE"
)

// GaxyError is a custom error type with context
type GaxyError struct {
	Type    ErrorType
	Message string
	Err     error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *GaxyError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap allows error unwrapping
func (e *GaxyError) Unwrap() error {
	return e.Err
}

// WithContext adds context to the error
func (e *GaxyError) WithContext(key string, value interface{}) *GaxyError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// New creates a new GaxyError
func New(errType ErrorType, message string) *GaxyError {
	return &GaxyError{
		Type:    errType,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with context
func Wrap(errType ErrorType, message string, err error) *GaxyError {
	return &GaxyError{
		Type:    errType,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// Common error constructors
func ConfigError(message string, err error) *GaxyError {
	return Wrap(ErrorTypeConfig, message, err)
}

func UpstreamError(message string, err error) *GaxyError {
	return Wrap(ErrorTypeUpstream, message, err)
}

func ProxyError(message string, err error) *GaxyError {
	return Wrap(ErrorTypeProxy, message, err)
}

func ValidationError(message string) *GaxyError {
	return New(ErrorTypeValidation, message)
}

func RateLimitError(message string) *GaxyError {
	return New(ErrorTypeRateLimit, message)
}

func InternalError(message string, err error) *GaxyError {
	return Wrap(ErrorTypeInternal, message, err)
}

func CacheError(message string, err error) *GaxyError {
	return Wrap(ErrorTypeCache, message, err)
}
