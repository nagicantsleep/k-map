package api

import (
	"encoding/json"
	"testing"
)

func TestGeocodeResult_JSONSerialization(t *testing.T) {
	t.Parallel()

	result := GeocodeResult{
		FormattedAddress: "1600 Amphitheatre Parkway, Mountain View, CA",
		Latitude:         37.422,
		Longitude:        -122.084,
		Confidence:       0.9,
		Source:           "osm",
		Components: AddressComponents{
			StreetNumber: "1600",
			Street:       "Amphitheatre Parkway",
			City:         "Mountain View",
			State:        "CA",
			PostalCode:   "94043",
			Country:      "United States",
			CountryCode:  "US",
		},
		PlaceType: "house",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled GeocodeResult
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.FormattedAddress != result.FormattedAddress {
		t.Errorf("expected formatted_address %s, got %s", result.FormattedAddress, unmarshaled.FormattedAddress)
	}

	if unmarshaled.Latitude != result.Latitude {
		t.Errorf("expected latitude %f, got %f", result.Latitude, unmarshaled.Latitude)
	}

	if unmarshaled.Confidence != result.Confidence {
		t.Errorf("expected confidence %f, got %f", result.Confidence, unmarshaled.Confidence)
	}

	if unmarshaled.Components.City != result.Components.City {
		t.Errorf("expected city %s, got %s", result.Components.City, unmarshaled.Components.City)
	}
}

func TestForwardGeocodeResponse_JSONSerialization(t *testing.T) {
	t.Parallel()

	resp := ForwardGeocodeResponse{
		Query: "1600 Amphitheatre Parkway",
		Results: []GeocodeResult{
			{
				FormattedAddress: "1600 Amphitheatre Parkway, Mountain View, CA",
				Latitude:         37.422,
				Longitude:        -122.084,
				Confidence:       0.9,
				Source:           "osm",
			},
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled ForwardGeocodeResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.Query != resp.Query {
		t.Errorf("expected query %s, got %s", resp.Query, unmarshaled.Query)
	}

	if len(unmarshaled.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(unmarshaled.Results))
	}

	if unmarshaled.Results[0].FormattedAddress != resp.Results[0].FormattedAddress {
		t.Errorf("expected formatted_address %s, got %s", resp.Results[0].FormattedAddress, unmarshaled.Results[0].FormattedAddress)
	}
}

func TestReverseGeocodeResponse_JSONSerialization(t *testing.T) {
	t.Parallel()

	resp := ReverseGeocodeResponse{
		Latitude:  37.422,
		Longitude: -122.084,
		Result: &GeocodeResult{
			FormattedAddress: "1600 Amphitheatre Parkway, Mountain View, CA",
			Latitude:         37.422,
			Longitude:        -122.084,
			Confidence:       0.9,
			Source:           "osm",
			PlaceType:        "house",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled ReverseGeocodeResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.Latitude != resp.Latitude {
		t.Errorf("expected latitude %f, got %f", resp.Latitude, unmarshaled.Latitude)
	}

	if unmarshaled.Result == nil {
		t.Fatal("expected result to not be nil")
	}

	if unmarshaled.Result.PlaceType != resp.Result.PlaceType {
		t.Errorf("expected place_type %s, got %s", resp.Result.PlaceType, unmarshaled.Result.PlaceType)
	}
}

func TestProximityResponse_JSONSerialization(t *testing.T) {
	t.Parallel()

	resp := ProximityResponse{
		IsNear:          true,
		DistanceMeters:  50.5,
		ThresholdMeters: 100,
		TargetMatch: &GeocodeResult{
			FormattedAddress: "1600 Amphitheatre Parkway, Mountain View, CA",
			Latitude:         37.422,
			Longitude:        -122.084,
			Confidence:       0.9,
			Source:           "osm",
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled ProximityResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.IsNear != resp.IsNear {
		t.Errorf("expected is_near %v, got %v", resp.IsNear, unmarshaled.IsNear)
	}

	if unmarshaled.DistanceMeters != resp.DistanceMeters {
		t.Errorf("expected distance_meters %f, got %f", resp.DistanceMeters, unmarshaled.DistanceMeters)
	}

	if unmarshaled.ThresholdMeters != resp.ThresholdMeters {
		t.Errorf("expected threshold_meters %f, got %f", resp.ThresholdMeters, unmarshaled.ThresholdMeters)
	}
}

func TestProximityResponse_NotNear(t *testing.T) {
	t.Parallel()

	resp := ProximityResponse{
		IsNear:          false,
		DistanceMeters:  1500.0,
		ThresholdMeters: 100,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled ProximityResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.IsNear {
		t.Error("expected is_near to be false")
	}

	if unmarshaled.TargetMatch != nil {
		t.Error("expected target_match to be nil")
	}
}

func TestForwardGeocodeRequest_JSONDeserialization(t *testing.T) {
	t.Parallel()

	jsonStr := `{"query":"1600 Amphitheatre Parkway","limit":5}`

	var req ForwardGeocodeRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Query != "1600 Amphitheatre Parkway" {
		t.Errorf("expected query '1600 Amphitheatre Parkway', got %s", req.Query)
	}

	if req.Limit != 5 {
		t.Errorf("expected limit 5, got %d", req.Limit)
	}
}

func TestReverseGeocodeRequest_JSONDeserialization(t *testing.T) {
	t.Parallel()

	jsonStr := `{"latitude":37.422,"longitude":-122.084}`

	var req ReverseGeocodeRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Latitude != 37.422 {
		t.Errorf("expected latitude 37.422, got %f", req.Latitude)
	}

	if req.Longitude != -122.084 {
		t.Errorf("expected longitude -122.084, got %f", req.Longitude)
	}
}

func TestProximityRequest_JSONDeserialization(t *testing.T) {
	t.Parallel()

	jsonStr := `{"latitude":37.42195,"longitude":-122.08405,"target_query":"1600 Amphitheatre Parkway","threshold_meters":100}`

	var req ProximityRequest
	if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if req.Latitude != 37.42195 {
		t.Errorf("expected latitude 37.42195, got %f", req.Latitude)
	}

	if req.TargetQuery != "1600 Amphitheatre Parkway" {
		t.Errorf("expected target_query '1600 Amphitheatre Parkway', got %s", req.TargetQuery)
	}

	if req.ThresholdMeters != 100 {
		t.Errorf("expected threshold_meters 100, got %f", req.ThresholdMeters)
	}
}
