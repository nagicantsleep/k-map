package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Validation errors.
var (
	ErrInvalidJSON       = errors.New("invalid JSON body")
	ErrMissingRequired   = errors.New("missing required field")
	ErrInvalidCoordinate = errors.New("invalid coordinate")
)

// DecodeJSON decodes a JSON request body and validates it.
func DecodeJSON(r *http.Request, dst interface{}) error {
	if r.Body == nil {
		return ErrInvalidJSON
	}
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		if errors.Is(err, io.EOF) {
			return ErrInvalidJSON
		}

		return fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	return nil
}

// ValidateCoordinate validates latitude and longitude ranges.
func ValidateCoordinate(lat, lng float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("%w: latitude must be between -90 and 90", ErrInvalidCoordinate)
	}

	if lng < -180 || lng > 180 {
		return fmt.Errorf("%w: longitude must be between -180 and 180", ErrInvalidCoordinate)
	}

	return nil
}

// ValidateRequiredString validates that a string field is not empty.
func ValidateRequiredString(value, fieldName string) error {
	if value == "" {
		return fmt.Errorf("%w: %s", ErrMissingRequired, fieldName)
	}

	return nil
}

// ValidationError represents a field-level validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors collects multiple validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "validation failed"
	}

	return fmt.Sprintf("validation failed: %v", []ValidationError(e))
}

// Add adds a validation error to the collection.
func (e *ValidationErrors) Add(field, message string) {
	*e = append(*e, ValidationError{Field: field, Message: message})
}

// HasErrors returns true if there are any validation errors.
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}
