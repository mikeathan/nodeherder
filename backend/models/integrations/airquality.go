package integrations

// Defines the provider-independent air-quality resource and its validation rules.

import (
	"math"
	"time"
)

// AirQualityCapability is the normalized air-quality resource contract.
const AirQualityCapability = "air_quality/v1"

// Supported air-quality index families.
const (
	AirQualityFamilyEuropean = "european"
	AirQualityFamilyUS       = "us"
)

// Server-derived air-quality categories.
const (
	AirQualityCategoryUnknown            = "unknown"
	AirQualityCategoryGood               = "good"
	AirQualityCategoryFair               = "fair"
	AirQualityCategoryModerate           = "moderate"
	AirQualityCategoryPoor               = "poor"
	AirQualityCategoryVeryPoor           = "very_poor"
	AirQualityCategoryExtremelyPoor      = "extremely_poor"
	AirQualityCategoryUnhealthySensitive = "unhealthy_sensitive"
	AirQualityCategoryUnhealthy          = "unhealthy"
	AirQualityCategoryVeryUnhealthy      = "very_unhealthy"
	AirQualityCategoryHazardous          = "hazardous"
)

// Read-time freshness values.
const (
	AirQualityFreshnessFresh = "fresh"
	AirQualityFreshnessStale = "stale"
)

// Resource bounds shared by validation and the capability codec.
const (
	AirQualityMaxHourly            = 168
	AirQualityMaxLocationLength    = 256
	AirQualityMaxAttributionLength = 512
	airQualityMaxIndex             = 1000
	airQualityMaxValue             = 10000
	airQualityFreshFor             = 6 * time.Hour
)

// AirQualityPollen carries optional pollen concentrations in grains per cubic metre.
type AirQualityPollen struct {
	Alder   *float64 `json:"alder"`
	Birch   *float64 `json:"birch"`
	Grass   *float64 `json:"grass"`
	Mugwort *float64 `json:"mugwort"`
	Olive   *float64 `json:"olive"`
	Ragweed *float64 `json:"ragweed"`
}

// AirQualityValues carries pollutant concentrations in micrograms per cubic metre.
type AirQualityValues struct {
	PM25            *float64 `json:"pm2_5"`
	PM10            *float64 `json:"pm10"`
	Ozone           *float64 `json:"ozone"`
	NitrogenDioxide *float64 `json:"nitrogen_dioxide"`
	SulphurDioxide  *float64 `json:"sulphur_dioxide"`
	CarbonMonoxide  *float64 `json:"carbon_monoxide"`
}

// AirQualityPoint is one timestamped forecast hour of pollutant values.
type AirQualityPoint struct {
	Time time.Time `json:"time"`
	AirQualityValues
}

// AirQualityForecast is the normalized, bounded air-quality snapshot.
// Latitude/Longitude are the selected setup origin when available. Dominant is empty when the
// provider does not identify one; it never guesses from incomparable pollutant units.
type AirQualityForecast struct {
	ContractVersion string            `json:"contract_version"`
	InstanceID      string            `json:"instance_id"`
	ResourceID      string            `json:"resource_id"`
	Revision        uint64            `json:"revision"`
	Location        string            `json:"location"`
	Timezone        string            `json:"timezone"`
	Latitude        *float64          `json:"latitude"`
	Longitude       *float64          `json:"longitude"`
	IndexFamily     string            `json:"index_family"`
	Index           *float64          `json:"index"`
	Category        string            `json:"category"`
	Dominant        string            `json:"dominant"`
	Current         *AirQualityValues `json:"current"`
	Hourly          []AirQualityPoint `json:"hourly"`
	Pollen          AirQualityPollen  `json:"pollen"`
	Attribution     string            `json:"attribution"`
	CapturedAt      time.Time         `json:"captured_at"`
	LastCheckedAt   time.Time         `json:"last_checked_at"`
	EvaluatedAt     time.Time         `json:"evaluated_at"`
	Freshness       string            `json:"freshness"`
}

// AirQualityCategoryFor derives the provider-independent category from the declared family.
// A nil or invalid index becomes unknown; the browser never invents thresholds.
func AirQualityCategoryFor(family string, index *float64) string {
	if index == nil || math.IsNaN(*index) || math.IsInf(*index, 0) || *index < 0 {
		return AirQualityCategoryUnknown
	}
	switch family {
	case AirQualityFamilyEuropean:
		switch {
		case *index <= 20:
			return AirQualityCategoryGood
		case *index <= 40:
			return AirQualityCategoryFair
		case *index <= 60:
			return AirQualityCategoryModerate
		case *index <= 80:
			return AirQualityCategoryPoor
		case *index <= 100:
			return AirQualityCategoryVeryPoor
		default:
			return AirQualityCategoryExtremelyPoor
		}
	case AirQualityFamilyUS:
		if *index > 500 {
			return AirQualityCategoryUnknown
		}
		switch {
		case *index <= 50:
			return AirQualityCategoryGood
		case *index <= 100:
			return AirQualityCategoryModerate
		case *index <= 150:
			return AirQualityCategoryUnhealthySensitive
		case *index <= 200:
			return AirQualityCategoryUnhealthy
		case *index <= 300:
			return AirQualityCategoryVeryUnhealthy
		default:
			return AirQualityCategoryHazardous
		}
	default:
		return AirQualityCategoryUnknown
	}
}

// Validate rejects invalid saved or normalized resources independently of provider decoding.
func (f AirQualityForecast) Validate() error {
	invalid := Fail("invalid_resource", "The air-quality data is invalid. Refresh or check the provider configuration.")
	if f.ContractVersion != AirQualityCapability || f.Revision == 0 {
		return invalid
	}
	for _, value := range []string{f.InstanceID, f.ResourceID, f.Location, f.Attribution} {
		if value == "" {
			return invalid
		}
	}
	if len(f.Location) > AirQualityMaxLocationLength || len(f.Attribution) > AirQualityMaxAttributionLength {
		return invalid
	}
	if f.Timezone == "" || f.Timezone == "Local" {
		return invalid
	}
	if _, err := time.LoadLocation(f.Timezone); err != nil {
		return invalid
	}
	if f.CapturedAt.IsZero() || f.LastCheckedAt.IsZero() {
		return invalid
	}
	if f.IndexFamily != AirQualityFamilyEuropean && f.IndexFamily != AirQualityFamilyUS {
		return invalid
	}
	if !validAirQualityCoordinates(f.Latitude, f.Longitude) {
		return invalid
	}
	if !validAirQualityIndex(f.Index) || f.Category != AirQualityCategoryFor(f.IndexFamily, f.Index) {
		return invalid
	}
	if f.Dominant != "" && !airQualityPollutant(f.Dominant) {
		return invalid
	}
	if f.Current != nil && !validAirQualityValues(*f.Current) {
		return invalid
	}
	if !validAirQualityPollen(f.Pollen) {
		return invalid
	}
	if len(f.Hourly) > AirQualityMaxHourly {
		return invalid
	}
	if (f.Current == nil || !hasAirQualityValue(*f.Current)) && len(f.Hourly) == 0 {
		return invalid
	}
	var previous time.Time
	for _, point := range f.Hourly {
		if !validAirQualityPoint(point) || !point.Time.After(previous) {
			return invalid
		}
		previous = point.Time
	}
	return nil
}

// validAirQualityPoint checks one forecast hour's timestamp and pollutant values.
func validAirQualityPoint(point AirQualityPoint) bool {
	return !point.Time.IsZero() && validAirQualityValues(point.AirQualityValues)
}

// validAirQualityValues rejects negative, non-finite or unbounded pollutant concentrations.
func validAirQualityValues(values AirQualityValues) bool {
	for _, value := range []*float64{
		values.PM25, values.PM10, values.Ozone, values.NitrogenDioxide,
		values.SulphurDioxide, values.CarbonMonoxide,
	} {
		if !validAirQualityValue(value) {
			return false
		}
	}
	return true
}

// hasAirQualityValue reports whether a current object carries at least one pollutant value.
func hasAirQualityValue(values AirQualityValues) bool {
	return values.PM25 != nil || values.PM10 != nil || values.Ozone != nil ||
		values.NitrogenDioxide != nil || values.SulphurDioxide != nil || values.CarbonMonoxide != nil
}

// validAirQualityPollen rejects negative, non-finite or unbounded pollen concentrations.
func validAirQualityPollen(pollen AirQualityPollen) bool {
	for _, value := range []*float64{
		pollen.Alder, pollen.Birch, pollen.Grass, pollen.Mugwort, pollen.Olive, pollen.Ragweed,
	} {
		if !validAirQualityValue(value) {
			return false
		}
	}
	return true
}

// validAirQualityIndex accepts nil or a finite non-negative bounded index.
func validAirQualityIndex(value *float64) bool {
	return value == nil || (!math.IsNaN(*value) && !math.IsInf(*value, 0) && *value >= 0 && *value <= airQualityMaxIndex)
}

// validAirQualityValue accepts nil or a finite non-negative bounded concentration.
func validAirQualityValue(value *float64) bool {
	return value == nil || (!math.IsNaN(*value) && !math.IsInf(*value, 0) && *value >= 0 && *value <= airQualityMaxValue)
}

// validAirQualityCoordinates requires both coordinates or neither and enforces WGS84 bounds.
func validAirQualityCoordinates(latitude, longitude *float64) bool {
	if latitude == nil && longitude == nil {
		return true
	}
	if latitude == nil || longitude == nil {
		return false
	}
	return !math.IsNaN(*latitude) && !math.IsInf(*latitude, 0) && *latitude >= -90 && *latitude <= 90 &&
		!math.IsNaN(*longitude) && !math.IsInf(*longitude, 0) && *longitude >= -180 && *longitude <= 180
}

// airQualityPollutant recognizes the pollutants a provider may name as dominant.
func airQualityPollutant(value string) bool {
	switch value {
	case "pm2_5", "pm10", "ozone", "nitrogen_dioxide", "sulphur_dioxide", "carbon_monoxide":
		return true
	default:
		return false
	}
}

// Project derives freshness without changing saved content, identity or provider evidence.
func (f AirQualityForecast) Project(now time.Time) AirQualityForecast {
	f.EvaluatedAt = now.UTC()
	f.Freshness = AirQualityFreshnessFresh
	if !now.Before(f.CapturedAt.Add(airQualityFreshFor)) {
		f.Freshness = AirQualityFreshnessStale
	}
	return f
}
