package integrations

import (
	"testing"
	"time"
)

func TestFuelPricesProjectionOrder(t *testing.T) {
	now := time.Date(2026, 10, 29, 14, 0, 0, 0, time.UTC)
	fuel := testFuel(now)
	projected := fuel.Project(now)
	if projected.Freshness != FuelFreshnessFresh {
		t.Fatalf("freshness = %q", projected.Freshness)
	}
	ordered := projected.Stations
	if len(ordered) != 4 {
		t.Fatalf("stations = %d", len(ordered))
	}
	if ordered[0].StationID != "near-no-price" || ordered[1].StationID != "cheap" || ordered[2].StationID != "pricey" || ordered[3].StationID != "no-location" {
		t.Fatalf("unexpected order: %s, %s, %s, %s", ordered[0].StationID, ordered[1].StationID, ordered[2].StationID, ordered[3].StationID)
	}
	if fuel.Stations[0].StationID != "pricey" {
		t.Fatal("Project mutated the saved stations")
	}
	if stale := fuel.Project(now.Add(time.Hour)); stale.Freshness != FuelFreshnessStale {
		t.Fatalf("stale freshness = %q", stale.Freshness)
	}
}

func TestFuelPricesValidation(t *testing.T) {
	now := time.Date(2026, 10, 29, 14, 0, 0, 0, time.UTC)
	fixture := func() FuelPrices { return testFuel(now) }
	if err := fixture().Validate(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		change func(*FuelPrices)
	}{
		{"bad radius", func(f *FuelPrices) { f.RadiusMetres = 50 }},
		{"bad currency", func(f *FuelPrices) { f.Currency = "EUR" }},
		{"missing origin", func(f *FuelPrices) { f.Origin.Label = "" }},
		{"duplicate station", func(f *FuelPrices) { f.Stations = append(f.Stations, f.Stations[0]) }},
		{"missing name", func(f *FuelPrices) { f.Stations[0].Name = "" }},
		{"single coordinate", func(f *FuelPrices) {
			latitude := 53.3
			f.Stations[0].Latitude = &latitude
			f.Stations[0].Longitude = nil
		}},
		{"negative distance", func(f *FuelPrices) { value := -1; f.Stations[0].DistanceMetres = &value }},
		{"zero price", func(f *FuelPrices) { f.Stations[0].Prices[0].PencePerLitre = 0 }},
		{"too high price", func(f *FuelPrices) { f.Stations[0].Prices[0].PencePerLitre = 1001 }},
		{"duplicate grade", func(f *FuelPrices) { f.Stations[0].Prices = append(f.Stations[0].Prices, f.Stations[0].Prices[0]) }},
		{"unordered grades", func(f *FuelPrices) {
			f.Stations[0].Prices = []FuelPrice{
				{Grade: FuelGradeB7, Label: "B7", PencePerLitre: 150},
				{Grade: FuelGradeE10, Label: "E10", PencePerLitre: 143},
			}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fuel := fixture()
			tt.change(&fuel)
			if fuel.Validate() == nil {
				t.Fatal("invalid fuel snapshot accepted")
			}
		})
	}
}

func TestOrderFuelPrices(t *testing.T) {
	prices := []FuelPrice{
		{Grade: FuelGradeOther, Label: "Other", PencePerLitre: 200},
		{Grade: FuelGradeB7, Label: "B7", PencePerLitre: 150},
		{Grade: FuelGradeE10, Label: "E10", PencePerLitre: 143},
	}
	ordered := OrderFuelPrices(prices)
	if ordered[0].Grade != FuelGradeE10 || ordered[1].Grade != FuelGradeB7 || ordered[2].Grade != FuelGradeOther {
		t.Fatalf("unexpected order: %+v", ordered)
	}
	if prices[0].Grade != FuelGradeOther {
		t.Fatal("OrderFuelPrices mutated its input")
	}
}

func testFuel(now time.Time) FuelPrices {
	near, cheap, pricey := 800, 1200, 1200
	return FuelPrices{
		ContractVersion: FuelPricesCapability,
		InstanceID:      "instance",
		ResourceID:      "resource",
		Revision:        1,
		Origin:          FuelOrigin{Label: "Nether Edge", Latitude: 53.3601, Longitude: -1.4832},
		RadiusMetres:    5000,
		Currency:        FuelCurrencyGBP,
		Stations: []FuelStation{
			{StationID: "pricey", Name: "Pricey", DistanceMetres: &pricey, Prices: []FuelPrice{{Grade: FuelGradeE10, Label: "E10", PencePerLitre: 143}}},
			{StationID: "cheap", Name: "Cheap", DistanceMetres: &cheap, Prices: []FuelPrice{{Grade: FuelGradeE10, Label: "E10", PencePerLitre: 141}}},
			{StationID: "near-no-price", Name: "Near", DistanceMetres: &near},
			{StationID: "no-location", Name: "Far"},
		},
		Attribution:   "Provider",
		CapturedAt:    now,
		LastCheckedAt: now,
	}
}
