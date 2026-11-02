package integrations

// Defines the provider-independent fuel-price resource, distance ordering and freshness.

import (
	"math"
	"slices"
	"sort"
	"time"
)

// FuelPricesCapability is the normalized fuel-price contract.
const FuelPricesCapability = "fuel_prices/v1"

// FuelCurrencyGBP is the only supported v1 currency.
const FuelCurrencyGBP = "GBP"

// Fuel grade identifiers.
type FuelGrade string

const (
	FuelGradeE10   FuelGrade = "e10"
	FuelGradeE5    FuelGrade = "e5"
	FuelGradeB7    FuelGrade = "b7"
	FuelGradeB10   FuelGrade = "b10"
	FuelGradeHVO   FuelGrade = "hvo"
	FuelGradeOther FuelGrade = "other"
)

// Read-time freshness values.
const (
	FuelFreshnessFresh = "fresh"
	FuelFreshnessStale = "stale"
)

// Resource bounds shared by validation and provider normalization.
const (
	FuelMinRadiusMetres = 100
	FuelMaxRadiusMetres = 50000
	FuelMaxStations     = 100
	FuelMaxPrices       = 6
	FuelMaxOriginLabel  = 256
	FuelMaxStationID    = 128
	FuelMaxStationName  = 160
	FuelMaxBrand        = 80
	FuelMaxAddress      = 240
	FuelMaxPostcode     = 16
	FuelMaxPriceLabel   = 40
	FuelMaxAttribution  = 512
	FuelMaxPricePence   = 1000
	fuelFreshFor        = time.Hour
)

// FuelOrigin is the selected search location; it is required and never silently changed on refresh.
type FuelOrigin struct {
	Label     string  `json:"label"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// FuelPrice is one grade's integer pence-per-litre price.
type FuelPrice struct {
	Grade         FuelGrade  `json:"grade"`
	Label         string     `json:"label"`
	PencePerLitre int        `json:"pence_per_litre"`
	UpdatedAt     *time.Time `json:"updated_at"`
}

// FuelStation is one forecourt with optional coordinates and distance from the origin.
type FuelStation struct {
	StationID      string      `json:"station_id"`
	Name           string      `json:"name"`
	Brand          string      `json:"brand"`
	Address        string      `json:"address"`
	Postcode       string      `json:"postcode"`
	Latitude       *float64    `json:"latitude"`
	Longitude      *float64    `json:"longitude"`
	DistanceMetres *int        `json:"distance_metres"`
	Prices         []FuelPrice `json:"prices"`
}

// FuelPrices is the normalized, bounded fuel-price snapshot.
type FuelPrices struct {
	ContractVersion string        `json:"contract_version"`
	InstanceID      string        `json:"instance_id"`
	ResourceID      string        `json:"resource_id"`
	Revision        uint64        `json:"revision"`
	Origin          FuelOrigin    `json:"origin"`
	RadiusMetres    int           `json:"radius_metres"`
	Currency        string        `json:"currency"`
	Stations        []FuelStation `json:"stations"`
	CapturedAt      time.Time     `json:"captured_at"`
	LastCheckedAt   time.Time     `json:"last_checked_at"`
	EvaluatedAt     time.Time     `json:"evaluated_at"`
	Freshness       string        `json:"freshness"`
	Attribution     string        `json:"attribution"`
}

// Validate rejects invalid saved or normalized fuel snapshots independently of provider decoding.
func (f FuelPrices) Validate() error {
	invalid := Fail("invalid_resource", "The fuel-price data is invalid. Refresh or check the provider configuration.")
	if f.ContractVersion != FuelPricesCapability || f.Revision == 0 {
		return invalid
	}
	for _, value := range []string{f.InstanceID, f.ResourceID, f.Attribution} {
		if value == "" {
			return invalid
		}
	}
	if len(f.Attribution) > FuelMaxAttribution || f.CapturedAt.IsZero() || f.LastCheckedAt.IsZero() {
		return invalid
	}
	if f.Origin.Label == "" || len(f.Origin.Label) > FuelMaxOriginLabel || !validCoordinates(f.Origin.Latitude, f.Origin.Longitude) {
		return invalid
	}
	if f.RadiusMetres < FuelMinRadiusMetres || f.RadiusMetres > FuelMaxRadiusMetres || f.Currency != FuelCurrencyGBP {
		return invalid
	}
	if len(f.Stations) > FuelMaxStations {
		return invalid
	}
	seen := make(map[string]bool, len(f.Stations))
	for _, station := range f.Stations {
		if !validStation(station, seen) {
			return invalid
		}
	}
	return nil
}

// validStation checks identity, text bounds, optional coordinates and price entries.
func validStation(station FuelStation, seen map[string]bool) bool {
	if station.StationID == "" || len(station.StationID) > FuelMaxStationID || seen[station.StationID] {
		return false
	}
	seen[station.StationID] = true
	if station.Name == "" || len(station.Name) > FuelMaxStationName ||
		len(station.Brand) > FuelMaxBrand || len(station.Address) > FuelMaxAddress ||
		len(station.Postcode) > FuelMaxPostcode {
		return false
	}
	if !validOptionalCoordinates(station.Latitude, station.Longitude) {
		return false
	}
	if station.DistanceMetres != nil && *station.DistanceMetres < 0 {
		return false
	}
	if len(station.Prices) > FuelMaxPrices {
		return false
	}
	grades := make(map[FuelGrade]bool, len(station.Prices))
	for index, price := range station.Prices {
		if !validFuelGrade(price.Grade) || price.Label == "" || len(price.Label) > FuelMaxPriceLabel ||
			price.PencePerLitre <= 0 || price.PencePerLitre > FuelMaxPricePence {
			return false
		}
		if price.UpdatedAt != nil && price.UpdatedAt.IsZero() {
			return false
		}
		if grades[price.Grade] {
			return false
		}
		grades[price.Grade] = true
		if index > 0 && fuelGradeRank(station.Prices[index-1].Grade) > fuelGradeRank(price.Grade) {
			return false
		}
	}
	return true
}

// validCoordinates requires both coordinates or neither and enforces WGS84 bounds.
func validCoordinates(latitude, longitude float64) bool {
	return !math.IsNaN(latitude) && !math.IsInf(latitude, 0) && latitude >= -90 && latitude <= 90 &&
		!math.IsNaN(longitude) && !math.IsInf(longitude, 0) && longitude >= -180 && longitude <= 180
}

// validCoordinates accepts optional station coordinates.
func validOptionalCoordinates(latitude, longitude *float64) bool {
	if latitude == nil && longitude == nil {
		return true
	}
	if latitude == nil || longitude == nil {
		return false
	}
	return validCoordinates(*latitude, *longitude)
}

// validFuelGrade checks the normalized grade enum.
func validFuelGrade(value FuelGrade) bool {
	switch value {
	case FuelGradeE10, FuelGradeE5, FuelGradeB7, FuelGradeB10, FuelGradeHVO, FuelGradeOther:
		return true
	default:
		return false
	}
}

// fuelGradeRank defines the display order inside one station.
func fuelGradeRank(value FuelGrade) int {
	switch value {
	case FuelGradeE10:
		return 0
	case FuelGradeE5:
		return 1
	case FuelGradeB7:
		return 2
	case FuelGradeB10:
		return 3
	case FuelGradeHVO:
		return 4
	default:
		return 5
	}
}

// OrderFuelPrices returns a display-ordered copy without mutating stored slices.
func OrderFuelPrices(prices []FuelPrice) []FuelPrice {
	ordered := slices.Clone(prices)
	sort.SliceStable(ordered, func(left, right int) bool {
		return fuelGradeRank(ordered[left].Grade) < fuelGradeRank(ordered[right].Grade)
	})
	return ordered
}

// Project derives freshness and distance ordering without changing content identity or capture age.
func (f FuelPrices) Project(now time.Time) FuelPrices {
	f.EvaluatedAt = now.UTC()
	f.Freshness = FuelFreshnessFresh
	if !now.Before(f.CapturedAt.Add(fuelFreshFor)) {
		f.Freshness = FuelFreshnessStale
	}
	f.Stations = orderFuelStations(f.Stations)
	return f
}

// orderFuelStations sorts a copy by distance, then by cheapest valid displayed grade.
func orderFuelStations(stations []FuelStation) []FuelStation {
	ordered := slices.Clone(stations)
	for index := range ordered {
		ordered[index].Prices = OrderFuelPrices(ordered[index].Prices)
	}
	sort.SliceStable(ordered, func(left, right int) bool {
		leftDistance, rightDistance := ordered[left].DistanceMetres, ordered[right].DistanceMetres
		if leftDistance == nil || rightDistance == nil {
			return leftDistance != nil
		}
		if *leftDistance != *rightDistance {
			return *leftDistance < *rightDistance
		}
		return cheapestPrice(ordered[left]) < cheapestPrice(ordered[right])
	})
	return ordered
}

// cheapestPrice returns the lowest valid price, or max int when a station has no prices.
func cheapestPrice(station FuelStation) int {
	if len(station.Prices) == 0 {
		return math.MaxInt
	}
	cheapest := math.MaxInt
	for _, price := range station.Prices {
		if price.PencePerLitre < cheapest {
			cheapest = price.PencePerLitre
		}
	}
	return cheapest
}
