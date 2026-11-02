package fuel

// Checks fuel-price mapping, origin distance, price filtering and projection.

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	defaults "node-herder/configs/integrations"
	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/definitions"
	model "node-herder/models/integrations"
)

const fuelPayload = `{
  "stations":[
    {
      "site_id":"a",
      "name":"Shell Abbeydale",
      "brand":"Shell",
      "address":"Abbeydale Road",
      "postcode":"S7 2QF",
      "lat":53.3552,
      "lon":-1.4921,
      "prices":[
        {"fuel":"E10","name":"E10","price":143,"updated":"2026-10-29T13:50:00Z"},
        {"fuel":"B7","name":"B7","price":150,"updated":"2026-10-29T13:50:00Z"},
        {"fuel":"E10","name":"Duplicate","price":140},
        {"fuel":"E5","name":"E5","price":0},
        {"fuel":"Mystery","name":"Super","price":200}
      ]
    },
    {
      "site_id":"b",
      "name":"No prices",
      "lat":53.36,
      "lon":-1.49,
      "prices":[]
    }
  ]
}`

func fuelMapping() mapping {
	return mapping{
		Attribution:  "Fuel prices: UK Fuel Finder.",
		Currency:     model.FuelCurrencyGBP,
		RadiusMetres: 5000,
		Stations: stationSeries{
			Items: "/stations",
			Fields: map[string]string{
				"station_id": "/site_id",
				"name":       "/name",
				"brand":      "/brand",
				"address":    "/address",
				"postcode":   "/postcode",
				"latitude":   "/lat",
				"longitude":  "/lon",
			},
			Prices: priceSeries{
				Items:  "/prices",
				Format: priceFormatRFC3339,
				Fields: map[string]string{
					"grade":           "/fuel",
					"label":           "/name",
					"pence_per_litre": "/price",
					"updated_at":      "/updated",
				},
			},
		},
	}
}

func normalizeFuel(t *testing.T, payload string) (model.FuelPrices, error) {
	t.Helper()
	var root any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(fuelMapping())
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err = json.Unmarshal(encoded, &raw); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 29, 14, 2, 0, 0, time.UTC)
	result, err := (Adapter{}).Normalize(root, raw, capabilities.Context{
		InstanceID:    "instance",
		ProviderID:    "fuel_finder",
		SelectedLabel: "Nether Edge",
		Origin:        &capabilities.Origin{Latitude: 53.3601, Longitude: -1.4832},
		Now:           now,
		NewIdentity: func(string) (string, error) {
			return "resource", nil
		},
	})
	if err != nil {
		return model.FuelPrices{}, err
	}
	var fuel model.FuelPrices
	if err = result.Resource.Decode(&fuel); err != nil {
		t.Fatal(err)
	}
	return fuel, nil
}

func TestShippedFuelProviderCompiles(t *testing.T) {
	source, err := defaults.Files.ReadFile("fuel_prices/fuel_finder.yaml")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := capabilities.NewRegistry(Adapter{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = definitions.NewCompiler(registry).CompileYAML(string(source)); err != nil {
		if failure, ok := err.(*model.Failure); ok {
			t.Fatalf("%+v", failure.Fields)
		}
		t.Fatal(err)
	}
}

func TestFuelJoinedPriceCollection(t *testing.T) {
	selectors := mapping{
		Attribution:  "Provider",
		Currency:     model.FuelCurrencyGBP,
		RadiusMetres: 50000,
		Stations: stationSeries{
			Items: "/stations",
			Fields: map[string]string{
				"station_id": "/node_id",
				"name":       "/trading_name",
				"brand":      "/brand_name",
				"address":    "/location/address_line_1",
				"postcode":   "/location/postcode",
				"latitude":   "/location/latitude",
				"longitude":  "/location/longitude",
			},
		},
		Prices: &priceList{
			Items:     "/prices",
			StationID: "/node_id",
			Prices:    "/fuel_prices",
			Format:    priceFormatLocal,
			Fields: map[string]string{
				"grade":           "/fuel_type",
				"pence_per_litre": "/price",
				"updated_at":      "/price_last_updated",
			},
		},
	}
	payload := `{
	  "stations":[{"node_id":"n1","trading_name":"Test","brand_name":"Test","location":{"address_line_1":"1 Road","postcode":"S1 1AA","latitude":"53.36","longitude":"-1.48"}}],
	  "prices":[{"node_id":"n1","fuel_prices":[{"fuel_type":"E10","price":"0120.0000","price_last_updated":"2026-01-12T08:15:23"}]}]
	}`
	var root any
	decoder := json.NewDecoder(strings.NewReader(payload))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(selectors)
	var raw map[string]any
	_ = json.Unmarshal(encoded, &raw)
	result, err := (Adapter{}).Normalize(root, raw, capabilities.Context{
		InstanceID:    "instance",
		SelectedLabel: "Nether Edge",
		Origin:        &capabilities.Origin{Latitude: 53.3601, Longitude: -1.4832},
		Now:           time.Date(2026, 1, 12, 10, 0, 0, 0, time.UTC),
		NewIdentity:   func(string) (string, error) { return "resource", nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	var fuel model.FuelPrices
	if err = result.Resource.Decode(&fuel); err != nil {
		t.Fatal(err)
	}
	if len(fuel.Stations) != 1 || len(fuel.Stations[0].Prices) != 1 || fuel.Stations[0].Prices[0].PencePerLitre != 120 {
		t.Fatalf("unexpected joined fuel: %+v", fuel.Stations)
	}
	if fuel.Stations[0].Prices[0].UpdatedAt == nil {
		t.Fatal("local timestamp not parsed")
	}
}

func TestFuelNormalization(t *testing.T) {
	fuel, err := normalizeFuel(t, fuelPayload)
	if err != nil {
		t.Fatal(err)
	}
	if fuel.Origin.Label != "Nether Edge" || fuel.Currency != model.FuelCurrencyGBP || fuel.RadiusMetres != 5000 {
		t.Fatalf("unexpected header: %+v", fuel)
	}
	if len(fuel.Stations) != 2 {
		t.Fatalf("stations = %d", len(fuel.Stations))
	}
	station := fuel.Stations[0]
	if station.StationID != "a" && station.StationID != "b" {
		t.Fatalf("unexpected station id: %s", station.StationID)
	}
	if station.DistanceMetres == nil || *station.DistanceMetres < 0 {
		t.Fatalf("missing distance: %+v", station)
	}
	var shell *model.FuelStation
	for index := range fuel.Stations {
		if fuel.Stations[index].StationID == "a" {
			shell = &fuel.Stations[index]
		}
	}
	if shell == nil {
		t.Fatal("Shell station missing")
	}
	if len(shell.Prices) != 3 {
		t.Fatalf("prices = %+v", shell.Prices)
	}
	if shell.Prices[0].Grade != model.FuelGradeE10 || shell.Prices[0].PencePerLitre != 143 ||
		shell.Prices[1].Grade != model.FuelGradeB7 || shell.Prices[2].Grade != model.FuelGradeOther {
		t.Fatalf("unexpected prices: %+v", shell.Prices)
	}
	if shell.Prices[0].UpdatedAt == nil {
		t.Fatal("price update time missing")
	}
	if fuel.Freshness != model.FuelFreshnessFresh {
		t.Fatalf("freshness = %q", fuel.Freshness)
	}
}

func TestFuelNormalizationRoundsProviderTenths(t *testing.T) {
	payload := strings.Replace(fuelPayload, `"price":143`, `"price":142.9`, 1)
	payload = strings.Replace(payload, `"price":140`, `"price":0`, 1)
	fuel, err := normalizeFuel(t, payload)
	if err != nil {
		t.Fatal(err)
	}
	for _, station := range fuel.Stations {
		for _, price := range station.Prices {
			if price.Grade == model.FuelGradeE10 {
				if price.PencePerLitre != 143 {
					t.Fatalf("rounded price = %+v", price)
				}
				return
			}
		}
	}
	t.Fatal("rounded E10 price missing")
}

func TestFuelNormalizationRejectsMissingNameAndOrigin(t *testing.T) {
	if _, err := normalizeFuel(t, strings.Replace(fuelPayload, `"name":"Shell Abbeydale"`, `"name":""`, 1)); err == nil {
		t.Fatal("missing station name accepted")
	}
	var root any
	decoder := json.NewDecoder(strings.NewReader(fuelPayload))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(fuelMapping())
	var raw map[string]any
	_ = json.Unmarshal(encoded, &raw)
	if _, err := (Adapter{}).Normalize(root, raw, capabilities.Context{
		InstanceID: "instance",
		Now:        time.Date(2026, 10, 29, 14, 2, 0, 0, time.UTC),
		NewIdentity: func(string) (string, error) {
			return "resource", nil
		},
	}); err == nil {
		t.Fatal("missing origin accepted")
	}
}

func TestFuelAdapterSourceAndProjection(t *testing.T) {
	fuel, err := normalizeFuel(t, fuelPayload)
	if err != nil {
		t.Fatal(err)
	}
	resource, err := model.NewResource(fuel)
	if err != nil {
		t.Fatal(err)
	}
	sources, err := (Adapter{}).Sources(resource)
	if err != nil || len(sources) != 1 || sources[0].Kind != "fuel_prices" || sources[0].ResourceID != "resource" {
		t.Fatalf("unexpected sources: %+v, %v", sources, err)
	}
	projected, err := (Adapter{}).Project(resource, time.Date(2026, 10, 29, 15, 2, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if err = projected.Decode(&fuel); err != nil || fuel.Freshness != model.FuelFreshnessStale {
		t.Fatalf("expected stale projection: %+v, %v", fuel, err)
	}
}
