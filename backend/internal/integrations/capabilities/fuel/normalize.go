package fuel

// Builds a validated fuel-price resource with origin-relative distance ordering.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
)

const (
	maxSourceStations    = operations.MaxPaginationItems
	maxPriceRows         = 12
	maxFuelPriceRows     = operations.MaxPaginationItems
	maxFuelUnixTimestamp = 7258118400
	earthRadiusMetres    = 6371000
	priceLocalLayout     = "2006-01-02T15:04:05"
)

// NormalizeFuel converts one provider result into a sorted, bounded fuel snapshot.
func NormalizeFuel(root any, selectors mapping, context capabilities.Context) (model.FuelPrices, error) {
	if context.Origin == nil {
		return model.FuelPrices{}, invalid("The provider setup did not record a search origin.")
	}
	label := limitText(clean(text(pointer(root, selectors.OriginLabel))), model.FuelMaxOriginLabel)
	if label == "" {
		label = limitText(clean(context.SelectedLabel), model.FuelMaxOriginLabel)
	}
	if label == "" {
		label = limitText(clean(context.SelectedValue), model.FuelMaxOriginLabel)
	}
	if label == "" {
		return model.FuelPrices{}, invalid("The provider omitted the origin label.")
	}
	fuel := model.FuelPrices{
		ContractVersion: model.FuelPricesCapability,
		InstanceID:      context.InstanceID,
		ResourceID:      context.ResourceID,
		Revision:        1,
		Origin: model.FuelOrigin{
			Label:     label,
			Latitude:  context.Origin.Latitude,
			Longitude: context.Origin.Longitude,
		},
		RadiusMetres:  selectors.RadiusMetres,
		Currency:      selectors.Currency,
		CapturedAt:    context.Now.UTC(),
		LastCheckedAt: context.Now.UTC(),
		Attribution:   selectors.Attribution,
	}
	if fuel.ResourceID == "" {
		id, err := context.NewIdentity("resource")
		if err != nil {
			return fuel, err
		}
		fuel.ResourceID = id
	}
	if context.Previous != nil {
		fuel.Revision = context.Previous.Revision + 1
	}
	items, ok := pointer(root, selectors.Stations.Items).([]any)
	if !ok || len(items) > maxSourceStations {
		return fuel, invalid("The provider returned an invalid station list.")
	}
	pricesByStation := map[string][]model.FuelPrice{}
	if selectors.Prices != nil {
		var err error
		pricesByStation, err = normalizePriceSeries(root, *selectors.Prices)
		if err != nil {
			return fuel, err
		}
	}
	fuel.Stations = make([]model.FuelStation, 0, len(items))
	for _, item := range items {
		station, err := normalizeStation(item, selectors, context, fuel.Origin, pricesByStation)
		if err != nil {
			return fuel, err
		}
		if station.DistanceMetres != nil && *station.DistanceMetres > fuel.RadiusMetres {
			continue
		}
		fuel.Stations = append(fuel.Stations, station)
	}
	fuel = fuel.Project(context.Now)
	if len(fuel.Stations) > model.FuelMaxStations {
		fuel.Stations = fuel.Stations[:model.FuelMaxStations]
	}
	return fuel, fuel.Validate()
}

// normalizeStation maps one forecourt and drops invalid individual prices, never the station.
func normalizeStation(
	item any,
	selectors mapping,
	context capabilities.Context,
	origin model.FuelOrigin,
	pricesByStation map[string][]model.FuelPrice,
) (model.FuelStation, error) {
	station := model.FuelStation{
		StationID: limitText(clean(text(pointer(item, selectors.Stations.Fields[string(stationFieldID)]))), model.FuelMaxStationID),
		Name:      limitText(clean(text(pointer(item, selectors.Stations.Fields[string(stationFieldName)]))), model.FuelMaxStationName),
		Brand:     limitText(clean(text(pointer(item, selectors.Stations.Fields[string(stationFieldBrand)]))), model.FuelMaxBrand),
		Address:   limitText(clean(text(pointer(item, selectors.Stations.Fields[string(stationFieldAddress)]))), model.FuelMaxAddress),
		Postcode:  limitText(clean(text(pointer(item, selectors.Stations.Fields[string(stationFieldPostcode)]))), model.FuelMaxPostcode),
	}
	if station.Name == "" {
		return station, invalid("The provider returned a station without a name.")
	}
	latitude := coordinate(pointer(item, selectors.Stations.Fields[string(stationFieldLatitude)]), 90)
	longitude := coordinate(pointer(item, selectors.Stations.Fields[string(stationFieldLongitude)]), 180)
	if latitude != nil && longitude != nil {
		station.Latitude = latitude
		station.Longitude = longitude
		distance := distanceMetres(origin.Latitude, origin.Longitude, *latitude, *longitude)
		station.DistanceMetres = &distance
	}
	if selectors.Prices != nil {
		station.Prices = dedupePrices(pricesByStation[station.StationID])
	} else {
		prices, err := normalizePrices(pointer(item, selectors.Stations.Prices.Items), selectors.Stations.Prices)
		if err != nil {
			return station, err
		}
		station.Prices = prices
	}
	if station.StationID == "" {
		station.StationID = derivedStationID(context.ProviderID, station, latitude, longitude)
	}
	return station, nil
}

// normalizePriceSeries joins a separate price collection to stations by provider station id.
func normalizePriceSeries(root any, selectors priceList) (map[string][]model.FuelPrice, error) {
	items, ok := pointer(root, selectors.Items).([]any)
	if !ok || len(items) > maxFuelPriceRows {
		return nil, invalid("The provider returned an invalid price collection.")
	}
	byStation := make(map[string][]model.FuelPrice)
	for _, item := range items {
		stationID := limitText(clean(text(pointer(item, selectors.StationID))), model.FuelMaxStationID)
		if stationID == "" {
			continue
		}
		if selectors.Prices != "" {
			nested, ok := pointer(item, selectors.Prices).([]any)
			if !ok {
				continue
			}
			for _, entry := range nested {
				if price, ok := normalizePrice(entry, selectors.Fields, selectors.Format); ok {
					byStation[stationID] = append(byStation[stationID], price)
				}
			}
			continue
		}
		if price, ok := normalizePrice(item, selectors.Fields, selectors.Format); ok {
			byStation[stationID] = append(byStation[stationID], price)
		}
	}
	for stationID, prices := range byStation {
		byStation[stationID] = dedupePrices(prices)
	}
	return byStation, nil
}

// normalizePrice builds one valid integer-pence entry, skipping invalid provider values.
func normalizePrice(item any, fields map[string]string, format priceTimeFormat) (model.FuelPrice, bool) {
	grade := fuelGrade(text(pointer(item, fields[string(priceFieldGrade)])))
	raw := number(pointer(item, fields[string(priceFieldPence)]))
	if raw == nil || math.IsNaN(*raw) || math.IsInf(*raw, 0) {
		return model.FuelPrice{}, false
	}
	pence := int(math.Round(*raw))
	if pence <= 0 || pence > model.FuelMaxPricePence {
		return model.FuelPrice{}, false
	}
	label := limitText(clean(text(pointer(item, fields[string(priceFieldLabel)]))), model.FuelMaxPriceLabel)
	if label == "" {
		label = gradeLabel(grade)
	}
	return model.FuelPrice{
		Grade:         grade,
		Label:         label,
		PencePerLitre: pence,
		UpdatedAt:     optionalInstant(pointer(item, fields[string(priceFieldUpdated)]), format),
	}, true
}

// dedupePrices keeps the first entry per grade and returns display order.
func dedupePrices(prices []model.FuelPrice) []model.FuelPrice {
	seen := make(map[model.FuelGrade]bool, len(prices))
	unique := make([]model.FuelPrice, 0, len(prices))
	for _, price := range prices {
		if seen[price.Grade] {
			continue
		}
		seen[price.Grade] = true
		unique = append(unique, price)
	}
	return model.OrderFuelPrices(unique)
}

// normalizePrices keeps only valid integer pence entries and deduplicates grades.
func normalizePrices(value any, selectors priceSeries) ([]model.FuelPrice, error) {
	if value == nil {
		return []model.FuelPrice{}, nil
	}
	items, ok := value.([]any)
	if !ok || len(items) > maxPriceRows {
		return nil, invalid("The provider returned an invalid price list.")
	}
	prices := make([]model.FuelPrice, 0, len(items))
	for _, item := range items {
		if price, ok := normalizePrice(item, selectors.Fields, selectors.Format); ok {
			prices = append(prices, price)
		}
	}
	return dedupePrices(prices), nil
}

// fuelGrade maps provider grade text to the normalized enum; unknown values stay other.
func fuelGrade(value string) model.FuelGrade {
	switch strings.ToLower(clean(value)) {
	case string(model.FuelGradeE10):
		return model.FuelGradeE10
	case string(model.FuelGradeE5):
		return model.FuelGradeE5
	case string(model.FuelGradeB7), "b7_standard", "b7_premium":
		return model.FuelGradeB7
	case string(model.FuelGradeB10):
		return model.FuelGradeB10
	case string(model.FuelGradeHVO):
		return model.FuelGradeHVO
	default:
		return model.FuelGradeOther
	}
}

// gradeLabel supplies a display label when the provider omits one.
func gradeLabel(grade model.FuelGrade) string {
	switch grade {
	case model.FuelGradeE10:
		return "E10"
	case model.FuelGradeE5:
		return "E5"
	case model.FuelGradeB7:
		return "B7"
	case model.FuelGradeB10:
		return "B10"
	case model.FuelGradeHVO:
		return "HVO"
	default:
		return "Fuel"
	}
}

// distanceMetres computes a rounded great-circle distance from the stored origin.
func distanceMetres(latitude1, longitude1, latitude2, longitude2 float64) int {
	radians := math.Pi / 180
	lat1, lon1, lat2, lon2 := latitude1*radians, longitude1*radians, latitude2*radians, longitude2*radians
	deltaLat, deltaLon := lat2-lat1, lon2-lon1
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	a = math.Max(0, math.Min(1, a))
	return int(math.Round(earthRadiusMetres * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))))
}

// derivedStationID builds a stable local id from the canonical forecourt tuple.
func derivedStationID(providerID string, station model.FuelStation, latitude, longitude *float64) string {
	coordinates := "none"
	if latitude != nil && longitude != nil {
		coordinates = fmt.Sprintf("%.5f,%.5f", *latitude, *longitude)
	}
	canonical := providerID + "\x00" + station.Name + "\x00" + station.Postcode + "\x00" + coordinates
	sum := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(sum[:16])
}

// coordinate returns a bounded optional latitude/longitude or nil when absent/invalid.
func coordinate(value any, limit float64) *float64 {
	result := number(value)
	if result == nil || math.IsNaN(*result) || math.IsInf(*result, 0) || *result < -limit || *result > limit {
		return nil
	}
	return result
}

// number accepts supported provider number representations.
func number(value any) *float64 {
	var result float64
	var err error
	switch typed := value.(type) {
	case json.Number:
		result, err = typed.Float64()
	case float64:
		result = typed
	case string:
		result, err = strconv.ParseFloat(typed, 64)
	default:
		return nil
	}
	if err != nil {
		return nil
	}
	return &result
}

// text converts supported provider scalar values into plain text.
func text(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case float64:
		return fmt.Sprintf("%g", typed)
	default:
		return ""
	}
}

// optionalInstant parses an optional timestamp and preserves absence/invalid values as nil.
func optionalInstant(value any, format priceTimeFormat) *time.Time {
	if value == nil {
		return nil
	}
	var parsed time.Time
	var err error
	if format == priceFormatUnix {
		n := number(value)
		if n == nil || math.Trunc(*n) != *n || *n < 0 || *n > maxFuelUnixTimestamp {
			return nil
		}
		parsed = time.Unix(int64(*n), 0).UTC()
	} else if format == priceFormatLocal {
		parsed, err = time.ParseInLocation(priceLocalLayout, text(value), time.UTC)
		if err != nil {
			return nil
		}
	} else {
		parsed, err = time.Parse(time.RFC3339, text(value))
		if err != nil {
			return nil
		}
	}
	parsed = parsed.UTC()
	return &parsed
}

// clean normalizes provider text whitespace.
func clean(value string) string { return strings.Join(strings.Fields(value), " ") }

// limitText bounds a string by rune count without splitting UTF-8.
func limitText(value string, max int) string {
	value = strings.TrimSpace(value)
	if len(value) <= max {
		return value
	}
	trimmed := []byte(value)[:max]
	for !utf8.Valid(trimmed) {
		trimmed = trimmed[:len(trimmed)-1]
	}
	return string(trimmed)
}

// pointer resolves one JSON pointer without exposing provider envelopes.
func pointer(root any, selector string) any {
	if selector == "" {
		return nil
	}
	value, _ := operations.Pointer(root, selector)
	return value
}

// invalid returns a sanitized provider response failure.
func invalid(message string) error { return model.Fail("invalid_response", message) }
