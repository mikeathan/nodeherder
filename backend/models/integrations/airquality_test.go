package integrations

import (
	"math"
	"testing"
	"time"
)

func TestAirQualityCategoryBands(t *testing.T) {
	tests := []struct {
		family string
		index  float64
		want   string
	}{
		{AirQualityFamilyEuropean, 20, AirQualityCategoryGood},
		{AirQualityFamilyEuropean, 21, AirQualityCategoryFair},
		{AirQualityFamilyEuropean, 40, AirQualityCategoryFair},
		{AirQualityFamilyEuropean, 41, AirQualityCategoryModerate},
		{AirQualityFamilyEuropean, 60, AirQualityCategoryModerate},
		{AirQualityFamilyEuropean, 61, AirQualityCategoryPoor},
		{AirQualityFamilyEuropean, 80, AirQualityCategoryPoor},
		{AirQualityFamilyEuropean, 81, AirQualityCategoryVeryPoor},
		{AirQualityFamilyEuropean, 100, AirQualityCategoryVeryPoor},
		{AirQualityFamilyEuropean, 101, AirQualityCategoryExtremelyPoor},
		{AirQualityFamilyUS, 50, AirQualityCategoryGood},
		{AirQualityFamilyUS, 51, AirQualityCategoryModerate},
		{AirQualityFamilyUS, 100, AirQualityCategoryModerate},
		{AirQualityFamilyUS, 101, AirQualityCategoryUnhealthySensitive},
		{AirQualityFamilyUS, 150, AirQualityCategoryUnhealthySensitive},
		{AirQualityFamilyUS, 151, AirQualityCategoryUnhealthy},
		{AirQualityFamilyUS, 200, AirQualityCategoryUnhealthy},
		{AirQualityFamilyUS, 201, AirQualityCategoryVeryUnhealthy},
		{AirQualityFamilyUS, 300, AirQualityCategoryVeryUnhealthy},
		{AirQualityFamilyUS, 301, AirQualityCategoryHazardous},
		{AirQualityFamilyUS, 501, AirQualityCategoryUnknown},
	}
	for _, tt := range tests {
		if got := AirQualityCategoryFor(tt.family, &tt.index); got != tt.want {
			t.Fatalf("%s/%v = %q, want %q", tt.family, tt.index, got, tt.want)
		}
	}
	if got := AirQualityCategoryFor(AirQualityFamilyEuropean, nil); got != AirQualityCategoryUnknown {
		t.Fatalf("nil index = %q", got)
	}
	if got := AirQualityCategoryFor("unknown", ptrFloat(10)); got != AirQualityCategoryUnknown {
		t.Fatalf("unknown family = %q", got)
	}
}

func TestAirQualityForecastValidation(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	fixture := func() AirQualityForecast {
		index, pollutant, pollen := 25.0, 12.0, 3.0
		current := AirQualityValues{PM25: &pollutant}
		hour := AirQualityPoint{Time: now, AirQualityValues: AirQualityValues{PM25: &pollutant}}
		return AirQualityForecast{
			ContractVersion: AirQualityCapability,
			InstanceID:      "instance",
			ResourceID:      "resource",
			Revision:        1,
			Location:        "Sheffield",
			Timezone:        "Europe/London",
			IndexFamily:     AirQualityFamilyEuropean,
			Index:           &index,
			Category:        AirQualityCategoryFor(AirQualityFamilyEuropean, &index),
			Current:         &current,
			Hourly:          []AirQualityPoint{hour},
			Pollen:          AirQualityPollen{Grass: &pollen},
			Attribution:     "Provider",
			CapturedAt:      now,
			LastCheckedAt:   now,
		}
	}
	if err := fixture().Validate(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		change func(*AirQualityForecast)
	}{
		{"unknown family", func(f *AirQualityForecast) { f.IndexFamily = "unknown" }},
		{"category mismatch", func(f *AirQualityForecast) { f.Category = AirQualityCategoryHazardous }},
		{"negative pollutant", func(f *AirQualityForecast) { f.Current.PM25 = ptrFloat(-1) }},
		{"invalid pollen", func(f *AirQualityForecast) { f.Pollen.Grass = ptrFloat(-1) }},
		{"unordered hours", func(f *AirQualityForecast) {
			next := f.Hourly[0]
			next.Time = f.Hourly[0].Time.Add(-time.Hour)
			f.Hourly = append(f.Hourly, next)
		}},
		{"single coordinate", func(f *AirQualityForecast) { f.Latitude = ptrFloat(53.3) }},
		{"missing attribution", func(f *AirQualityForecast) { f.Attribution = "" }},
		{"empty payload", func(f *AirQualityForecast) { f.Current = nil; f.Hourly = nil }},
		{"invalid dominant", func(f *AirQualityForecast) { f.Dominant = "mystery" }},
		{"non finite index", func(f *AirQualityForecast) { f.Index = ptrFloat(math.NaN()) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture()
			tt.change(&f)
			if f.Validate() == nil {
				t.Fatal("invalid air-quality forecast accepted")
			}
		})
	}
}

func TestAirQualityProjection(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	index := 25.0
	f := AirQualityForecast{
		ContractVersion: AirQualityCapability,
		InstanceID:      "instance",
		ResourceID:      "resource",
		Revision:        1,
		Location:        "Sheffield",
		Timezone:        "Europe/London",
		IndexFamily:     AirQualityFamilyEuropean,
		Index:           &index,
		Category:        AirQualityCategoryFor(AirQualityFamilyEuropean, &index),
		Attribution:     "Provider",
		CapturedAt:      now,
		LastCheckedAt:   now,
	}
	fresh := f.Project(now.Add(5 * time.Hour))
	if fresh.Freshness != AirQualityFreshnessFresh {
		t.Fatalf("fresh = %q", fresh.Freshness)
	}
	stale := f.Project(now.Add(6 * time.Hour))
	if stale.Freshness != AirQualityFreshnessStale {
		t.Fatalf("stale = %q", stale.Freshness)
	}
	if f.Freshness != "" || !f.EvaluatedAt.IsZero() {
		t.Fatal("projection mutated the saved forecast")
	}
}

func ptrFloat(value float64) *float64 { return &value }
