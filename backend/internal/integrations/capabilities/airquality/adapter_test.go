package airquality

// Checks the shipped air-quality mapping and normalization.

import (
	"encoding/json"
	defaults "node-herder/configs/integrations"
	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/definitions"
	model "node-herder/models/integrations"
	"strings"
	"testing"
	"time"
)

const openMeteoPayload = `{
  "timezone":"Europe/London",
  "current":{"time":1793188800,"european_aqi":25,"pm2_5":12,"pm10":20,"ozone":30,"nitrogen_dioxide":10,"sulphur_dioxide":2,"carbon_monoxide":200,"alder_pollen":1,"birch_pollen":2,"grass_pollen":3,"mugwort_pollen":0,"olive_pollen":null,"ragweed_pollen":null},
  "hourly":{"time":[1793188800,1793192400],"european_aqi":[25,35],"pm2_5":[12,15],"pm10":[20,24],"ozone":[30,32],"nitrogen_dioxide":[10,11],"sulphur_dioxide":[2,2],"carbon_monoxide":[200,210],"alder_pollen":[1,1],"birch_pollen":[2,2],"grass_pollen":[3,4],"mugwort_pollen":[0,0],"olive_pollen":[null,null],"ragweed_pollen":[null,null]}
}`

func TestOpenMeteoAirQualityNormalization(t *testing.T) {
	source, err := defaults.Files.ReadFile("air_quality/open_meteo.yaml")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := capabilities.NewRegistry(Adapter{})
	if err != nil {
		t.Fatal(err)
	}
	compiled, err := definitions.NewCompiler(registry).CompileYAML(string(source))
	if err != nil {
		t.Fatal(err)
	}
	var root any
	decoder := json.NewDecoder(strings.NewReader(openMeteoPayload))
	decoder.UseNumber()
	if err = decoder.Decode(&root); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	result, err := registry.Normalize(compiled.Definition, root, capabilities.Context{
		InstanceID:    "instance",
		SelectedLabel: "Sheffield",
		Origin:        &capabilities.Origin{Latitude: 53.3, Longitude: -1.4},
		Now:           now,
		NewIdentity:   func(string) (string, error) { return "resource", nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	var forecast model.AirQualityForecast
	if err = result.Resource.Decode(&forecast); err != nil {
		t.Fatal(err)
	}
	if forecast.Location != "Sheffield" || forecast.IndexFamily != model.AirQualityFamilyEuropean || forecast.Index == nil || *forecast.Index != 25 || forecast.Category != model.AirQualityCategoryFair {
		t.Fatalf("unexpected header: %+v", forecast)
	}
	if forecast.Latitude == nil || *forecast.Latitude != 53.3 || forecast.Longitude == nil || *forecast.Longitude != -1.4 {
		t.Fatalf("missing origin: %+v", forecast)
	}
	if forecast.Current == nil || forecast.Current.PM25 == nil || *forecast.Current.PM25 != 12 || forecast.Pollen.Grass == nil || *forecast.Pollen.Grass != 3 {
		t.Fatalf("missing current values: %+v / %+v", forecast.Current, forecast.Pollen)
	}
	if len(forecast.Hourly) != 2 || forecast.Hourly[1].PM25 == nil || *forecast.Hourly[1].PM25 != 15 {
		t.Fatalf("unexpected hourly values: %+v", forecast.Hourly)
	}
	if forecast.Freshness != model.AirQualityFreshnessFresh {
		t.Fatalf("freshness = %q", forecast.Freshness)
	}
	sources, err := Adapter{}.Sources(result.Resource)
	if err != nil || len(sources) != 1 || sources[0].Kind != "air_quality" || sources[0].ResourceID != "resource" {
		t.Fatalf("unexpected sources: %+v, %v", sources, err)
	}
	projected, err := (Adapter{}).Project(result.Resource, now.Add(6*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if err = projected.Decode(&forecast); err != nil || forecast.Freshness != model.AirQualityFreshnessStale {
		t.Fatalf("expected stale projection: %+v, %v", forecast, err)
	}
}

func TestOpenMeteoAirQualityRejectsInvalidValues(t *testing.T) {
	source, err := defaults.Files.ReadFile("air_quality/open_meteo.yaml")
	if err != nil {
		t.Fatal(err)
	}
	registry, err := capabilities.NewRegistry(Adapter{})
	if err != nil {
		t.Fatal(err)
	}
	compiled, err := definitions.NewCompiler(registry).CompileYAML(string(source))
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []struct{ old, replacement string }{
		{`"pm2_5":12`, `"pm2_5":-1`},
		{`"grass_pollen":[3,4]`, `"grass_pollen":[3]`},
		{`"european_aqi":25`, `"european_aqi":1001`},
		{`"time":[1793188800,1793192400]`, `"time":[1793016000,1793019600]`},
	} {
		var root any
		decoder := json.NewDecoder(strings.NewReader(strings.Replace(openMeteoPayload, mutation.old, mutation.replacement, 1)))
		decoder.UseNumber()
		if err = decoder.Decode(&root); err != nil {
			t.Fatal(err)
		}
		if _, err = registry.Normalize(compiled.Definition, root, capabilities.Context{
			InstanceID:    "instance",
			SelectedLabel: "Sheffield",
			Now:           time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC),
			NewIdentity: func(string) (string, error) {
				return "resource", nil
			},
		}); err == nil {
			t.Fatalf("accepted invalid payload %q", mutation.replacement)
		}
	}
}
