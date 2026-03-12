package api

// AddressComponents represents normalized address parts.
type AddressComponents struct {
	StreetNumber string `json:"street_number,omitempty"`
	Street       string `json:"street,omitempty"`
	City         string `json:"city,omitempty"`
	State        string `json:"state,omitempty"`
	PostalCode   string `json:"postal_code,omitempty"`
	Country      string `json:"country,omitempty"`
	CountryCode  string `json:"country_code,omitempty"`
}

// GeocodeResult represents a single geocoding result.
type GeocodeResult struct {
	FormattedAddress string           `json:"formatted_address"`
	Latitude         float64          `json:"latitude"`
	Longitude        float64          `json:"longitude"`
	Confidence       float64          `json:"confidence"`
	Source           string           `json:"source"`
	Components       AddressComponents `json:"components,omitempty"`
	PlaceType        string           `json:"place_type,omitempty"`
}

// ForwardGeocodeRequest represents a forward geocoding request.
type ForwardGeocodeRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// ForwardGeocodeResponse represents a forward geocoding response.
type ForwardGeocodeResponse struct {
	Query   string          `json:"query"`
	Results []GeocodeResult `json:"results"`
}

// ReverseGeocodeRequest represents a reverse geocoding request.
type ReverseGeocodeRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// ReverseGeocodeResponse represents a reverse geocoding response.
type ReverseGeocodeResponse struct {
	Latitude  float64       `json:"latitude"`
	Longitude float64       `json:"longitude"`
	Result    *GeocodeResult `json:"result,omitempty"`
}

// ProximityRequest represents a proximity validation request.
type ProximityRequest struct {
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	TargetQuery     string  `json:"target_query"`
	ThresholdMeters float64 `json:"threshold_meters"`
}

// ProximityResponse represents a proximity validation response.
type ProximityResponse struct {
	IsNear          bool            `json:"is_near"`
	DistanceMeters  float64         `json:"distance_meters"`
	ThresholdMeters float64         `json:"threshold_meters"`
	TargetMatch     *GeocodeResult  `json:"target_match,omitempty"`
}
