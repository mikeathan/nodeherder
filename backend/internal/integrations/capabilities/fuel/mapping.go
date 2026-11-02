package fuel

// Validates fuel-price provider selectors and response shapes.

import (
	"bytes"
	"encoding/json"

	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

// stationField identifies one forecourt property.
type stationField string

const (
	stationFieldID        stationField = "station_id"
	stationFieldName      stationField = "name"
	stationFieldBrand     stationField = "brand"
	stationFieldAddress   stationField = "address"
	stationFieldPostcode  stationField = "postcode"
	stationFieldLatitude  stationField = "latitude"
	stationFieldLongitude stationField = "longitude"
)

// priceField identifies one grade price property.
type priceField string

const (
	priceFieldStationID priceField = "station_id"
	priceFieldGrade     priceField = "grade"
	priceFieldLabel     priceField = "label"
	priceFieldPence     priceField = "pence_per_litre"
	priceFieldUpdated   priceField = "updated_at"
)

// priceTimeFormat identifies supported price timestamp encodings.
type priceTimeFormat string

const (
	priceFormatRFC3339 priceTimeFormat = "rfc3339"
	priceFormatUnix    priceTimeFormat = "unix"
	priceFormatLocal   priceTimeFormat = "local"
)

// priceSeries describes how one station's prices are shaped.
type priceSeries struct {
	Items  string            `json:"items"`
	Format priceTimeFormat   `json:"format"`
	Fields map[string]string `json:"fields"`
}

// stationSeries describes how the provider's forecourt list is shaped.
type stationSeries struct {
	Items  string            `json:"items"`
	Fields map[string]string `json:"fields"`
	Prices priceSeries       `json:"prices,omitempty"`
}

// priceList describes a separate price collection joined to stations by provider id.
type priceList struct {
	Items     string            `json:"items"`
	StationID string            `json:"station_id"`
	Prices    string            `json:"prices,omitempty"`
	Format    priceTimeFormat   `json:"format"`
	Fields    map[string]string `json:"fields"`
}

// mapping is the capability-owned provider recipe for fuel prices.
type mapping struct {
	Attribution  string        `json:"attribution"`
	Currency     string        `json:"currency"`
	RadiusMetres int           `json:"radius_metres"`
	OriginLabel  string        `json:"origin_label,omitempty"`
	Stations     stationSeries `json:"stations"`
	Prices       *priceList    `json:"prices,omitempty"`
}

// decode strictly parses a fuel mapping and validates its selectors.
func decode(raw map[string]any) (mapping, error) {
	var m mapping
	encoded, err := json.Marshal(raw)
	if err != nil {
		return m, err
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&m); err != nil {
		return m, model.InvalidDefinition("mapping", "Invalid fuel-price mapping.")
	}
	if m.Attribution == "" || len(m.Attribution) > model.FuelMaxAttribution {
		return m, model.InvalidDefinition("mapping.attribution", "Declare provider attribution.")
	}
	if m.Currency != model.FuelCurrencyGBP {
		return m, model.InvalidDefinition("mapping.currency", "Fuel prices must declare GBP.")
	}
	if m.RadiusMetres < model.FuelMinRadiusMetres || m.RadiusMetres > model.FuelMaxRadiusMetres {
		return m, model.InvalidDefinition("mapping.radius_metres", "Declare a bounded search radius.")
	}
	if m.OriginLabel != "" && !operations.ValidPointer(m.OriginLabel) {
		return m, model.InvalidDefinition("mapping.origin_label", "Use a JSON pointer for the origin label.")
	}
	if m.Stations.Items == "" || !operations.ValidPointer(m.Stations.Items) {
		return m, model.InvalidDefinition("mapping.stations.items", "Use a JSON pointer for station items.")
	}
	for name, selector := range m.Stations.Fields {
		if !supportedStationField(stationField(name)) || selector == "" || !operations.ValidPointer(selector) {
			return m, model.InvalidDefinition("mapping.stations.fields", "Invalid station selector.")
		}
	}
	if _, ok := m.Stations.Fields[string(stationFieldName)]; !ok {
		return m, model.InvalidDefinition("mapping.stations.fields", "Declare a station name selector.")
	}
	if m.Prices == nil {
		if err := validatePriceSeries("mapping.stations.prices", m.Stations.Prices); err != nil {
			return m, err
		}
		return m, nil
	}
	if _, ok := m.Stations.Fields[string(stationFieldID)]; !ok {
		return m, model.InvalidDefinition("mapping.stations.fields", "Declare a station id selector for the price join.")
	}
	if m.Prices.Items == "" || !operations.ValidPointer(m.Prices.Items) {
		return m, model.InvalidDefinition("mapping.prices.items", "Use a JSON pointer for price items.")
	}
	if m.Prices.StationID == "" || !operations.ValidPointer(m.Prices.StationID) {
		return m, model.InvalidDefinition("mapping.prices.station_id", "Use a JSON pointer for the price station id.")
	}
	if m.Prices.Prices != "" && !operations.ValidPointer(m.Prices.Prices) {
		return m, model.InvalidDefinition("mapping.prices.prices", "Use a JSON pointer for nested price entries.")
	}
	switch m.Prices.Format {
	case priceFormatRFC3339, priceFormatUnix, priceFormatLocal:
	default:
		return m, model.InvalidDefinition("mapping.prices.format", "Declare a supported price time format.")
	}
	for name, selector := range m.Prices.Fields {
		if !supportedPriceField(priceField(name)) || selector == "" || !operations.ValidPointer(selector) {
			return m, model.InvalidDefinition("mapping.prices.fields", "Invalid price selector.")
		}
	}
	for _, required := range []priceField{priceFieldGrade, priceFieldPence} {
		if _, ok := m.Prices.Fields[string(required)]; !ok {
			return m, model.InvalidDefinition("mapping.prices.fields", "Declare required price fields.")
		}
	}
	return m, nil
}

// validatePriceSeries checks either a nested station price series or a joined price collection.
func validatePriceSeries(path string, selectors priceSeries) error {
	if selectors.Items == "" || !operations.ValidPointer(selectors.Items) {
		return model.InvalidDefinition(path+".items", "Use a JSON pointer for price items.")
	}
	switch selectors.Format {
	case priceFormatRFC3339, priceFormatUnix, priceFormatLocal:
	default:
		return model.InvalidDefinition(path+".format", "Declare a supported price time format.")
	}
	for name, selector := range selectors.Fields {
		if !supportedPriceField(priceField(name)) || selector == "" || !operations.ValidPointer(selector) {
			return model.InvalidDefinition(path+".fields", "Invalid price selector.")
		}
	}
	for _, required := range []priceField{priceFieldGrade, priceFieldPence} {
		if _, ok := selectors.Fields[string(required)]; !ok {
			return model.InvalidDefinition(path+".fields", "Declare required price fields.")
		}
	}
	return nil
}

// supportedStationField limits mappings to normalized forecourt properties.
func supportedStationField(value stationField) bool {
	switch value {
	case stationFieldID, stationFieldName, stationFieldBrand, stationFieldAddress,
		stationFieldPostcode, stationFieldLatitude, stationFieldLongitude:
		return true
	default:
		return false
	}
}

// supportedPriceField limits mappings to normalized price properties.
func supportedPriceField(value priceField) bool {
	switch value {
	case priceFieldStationID, priceFieldGrade, priceFieldLabel, priceFieldPence, priceFieldUpdated:
		return true
	default:
		return false
	}
}
