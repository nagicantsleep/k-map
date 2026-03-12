package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON_ValidBody(t *testing.T) {
	t.Parallel()

	type testRequest struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	body := `{"name":"test","value":42}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))

	var dst testRequest
	err := DecodeJSON(req, &dst)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if dst.Name != "test" {
		t.Errorf("expected name 'test', got %s", dst.Name)
	}

	if dst.Value != 42 {
		t.Errorf("expected value 42, got %d", dst.Value)
	}
}

func TestDecodeJSON_EmptyBody(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/test", nil)

	var dst struct{}
	err := DecodeJSON(req, &dst)
	if err == nil {
		t.Fatal("expected error for nil body")
	}

	if !errors.Is(err, ErrInvalidJSON) {
		t.Errorf("expected ErrInvalidJSON, got %v", err)
	}
}

func TestDecodeJSON_InvalidJSON(t *testing.T) {
	t.Parallel()

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(body))

	var dst struct{}
	err := DecodeJSON(req, &dst)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}

	if !errors.Is(err, ErrInvalidJSON) {
		t.Errorf("expected ErrInvalidJSON, got %v", err)
	}
}

func TestValidateCoordinate_Valid(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		lat float64
		lng float64
	}{
		{0, 0},
		{90, 180},
		{-90, -180},
		{37.422, -122.084},
	}

	for _, tc := range testCases {
		err := ValidateCoordinate(tc.lat, tc.lng)
		if err != nil {
			t.Errorf("expected valid for lat=%f, lng=%f, got error: %v", tc.lat, tc.lng, err)
		}
	}
}

func TestValidateCoordinate_InvalidLatitude(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		lat float64
		lng float64
	}{
		{91, 0},
		{-91, 0},
		{100, 0},
		{-100, 0},
	}

	for _, tc := range testCases {
		err := ValidateCoordinate(tc.lat, tc.lng)
		if err == nil {
			t.Errorf("expected error for invalid latitude %f", tc.lat)
		}

		if !errors.Is(err, ErrInvalidCoordinate) {
			t.Errorf("expected ErrInvalidCoordinate, got %v", err)
		}
	}
}

func TestValidateCoordinate_InvalidLongitude(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		lat float64
		lng float64
	}{
		{0, 181},
		{0, -181},
		{0, 200},
		{0, -200},
	}

	for _, tc := range testCases {
		err := ValidateCoordinate(tc.lat, tc.lng)
		if err == nil {
			t.Errorf("expected error for invalid longitude %f", tc.lng)
		}

		if !errors.Is(err, ErrInvalidCoordinate) {
			t.Errorf("expected ErrInvalidCoordinate, got %v", err)
		}
	}
}

func TestValidateRequiredString_Valid(t *testing.T) {
	t.Parallel()

	err := ValidateRequiredString("value", "field")
	if err != nil {
		t.Errorf("expected no error for non-empty string, got %v", err)
	}
}

func TestValidateRequiredString_Empty(t *testing.T) {
	t.Parallel()

	err := ValidateRequiredString("", "field")
	if err == nil {
		t.Fatal("expected error for empty string")
	}

	if !errors.Is(err, ErrMissingRequired) {
		t.Errorf("expected ErrMissingRequired, got %v", err)
	}
}

func TestValidationErrors(t *testing.T) {
	t.Parallel()

	var errs ValidationErrors
	if errs.HasErrors() {
		t.Error("expected no errors initially")
	}

	errs.Add("field1", "is required")
	errs.Add("field2", "is invalid")

	if !errs.HasErrors() {
		t.Error("expected errors after Add")
	}

	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d", len(errs))
	}

	if errs[0].Field != "field1" {
		t.Errorf("expected field1, got %s", errs[0].Field)
	}

	if errs[1].Field != "field2" {
		t.Errorf("expected field2, got %s", errs[1].Field)
	}

	errMsg := errs.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}
