package weather

// Checks weather provider mappings and forecast normalization.

import (
	"encoding/json"
	"math"
	defaults "node-herder/configs/integrations"
	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/definitions"
	model "node-herder/models/integrations"
	"strings"
	"testing"
	"time"
)

func TestShippedProviderNormalization(t *testing.T) {
	tests := []struct {
		name, payload string
		wantWind      float64
	}{
		{"open_meteo", `{"utc_offset_seconds":0,"timezone":"Europe/London","current":{"time":1793188800,"temperature_2m":14,"apparent_temperature":12,"wind_speed_10m":18,"wind_gusts_10m":25,"wind_direction_10m":200,"relative_humidity_2m":72,"precipitation":0,"weather_code":2},"hourly":{"time":[1793188800],"temperature_2m":[14],"apparent_temperature":[12],"wind_speed_10m":[18],"wind_gusts_10m":[25],"wind_direction_10m":[200],"relative_humidity_2m":[72],"precipitation":[0],"weather_code":[2],"precipitation_probability":[20],"uv_index":[3]},"daily":{"time":[1793145600],"temperature_2m_max":[16],"temperature_2m_min":[9],"apparent_temperature_max":[14],"apparent_temperature_min":[7],"weather_code":[2],"wind_speed_10m_max":[18],"wind_gusts_10m_max":[29],"wind_direction_10m_dominant":[210],"precipitation_probability_max":[70],"precipitation_sum":[2],"sunrise":[1793171520],"sunset":[1793209680],"uv_index_max":[4]}}`, 18},
		{"weatherapi", `{"location":{"name":"Bristol","tz_id":"Europe/London"},"current":{"last_updated_epoch":1793188800,"temp_c":14,"condition":{"code":1003}},"forecast":{"forecastday":[{"date":"2026-10-28","day":{"maxtemp_c":16,"mintemp_c":9,"condition":{"code":1003},"maxwind_kph":18,"avghumidity":70,"daily_chance_of_rain":40,"daily_chance_of_snow":5,"totalprecip_mm":1.2,"uv":4},"hour":[{"time_epoch":1793188800,"temp_c":14,"wind_kph":18,"condition":{"code":1003}}]}]}}`, 18},
		{"met_office", `{"primary":{"features":[{"properties":{"location":{"name":"Bristol"},"timeSeries":[{"time":"2026-10-28T12:00Z","screenTemperature":14,"windSpeed10m":5,"significantWeatherCode":3,"uvIndex":2}]}}]},"additional":{"features":[{"properties":{"timeSeries":[{"time":"2026-10-28T00:00Z","dayMaxScreenTemperature":16,"nightMinScreenTemperature":9,"daySignificantWeatherCode":3,"nightSignificantWeatherCode":7,"dayMaxFeelsLikeTemp":14,"nightMinFeelsLikeTemp":7,"midday10MWindSpeed":5,"midday10MWindGust":8,"midday10MWindDirection":220,"middayRelativeHumidity":75,"middayMslp":101300,"middayVisibility":20000,"dayProbabilityOfPrecipitation":70,"dayProbabilityOfRain":60,"dayProbabilityOfSnow":0,"maxUvIndex":3}]}}]}}`, 18},
	}
	registry, _ := capabilities.NewRegistry(Adapter{})
	compiler := definitions.NewCompiler(registry)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, e := defaults.Files.ReadFile("weather/" + tt.name + ".yaml")
			if e != nil {
				t.Fatal(e)
			}
			compiled, e := compiler.CompileYAML(string(source))
			if e != nil {
				t.Fatal(e)
			}
			var root any
			decoder := json.NewDecoder(strings.NewReader(tt.payload))
			decoder.UseNumber()
			if e = decoder.Decode(&root); e != nil {
				t.Fatal(e)
			}
			context := capabilities.Context{InstanceID: "instance", SelectedValue: "51.45,-2.59", Now: time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC), NewIdentity: func(string) (string, error) { return "resource", nil }}
			result, e := registry.Normalize(compiled.Definition, root, context)
			if e != nil {
				t.Fatal(e)
			}
			var f model.WeatherForecast
			if e = result.Resource.Decode(&f); e != nil {
				t.Fatal(e)
			}
			if len(f.Hourly) != 1 || len(f.Daily) != 1 || f.Hourly[0].Temperature == nil || *f.Hourly[0].Temperature != 14 || f.Hourly[0].Wind == nil || *f.Hourly[0].Wind != tt.wantWind {
				t.Fatalf("invalid normalization: %+v", f)
			}
			if tt.name == "open_meteo" {
				day := f.Daily[0]
				if f.Hourly[0].UV == nil || *f.Hourly[0].UV != 3 || day.UVMax == nil || *day.UVMax != 4 || day.Sunrise == nil || *day.Sunrise != "07:12" || day.Sunset == nil || *day.Sunset != "17:48" {
					t.Fatalf("missing daily signals: %+v", day)
				}
				for _, mutation := range []struct{ old, replacement string }{
					{`"sunrise":[1793171520]`, `"sunrise":["7:12"]`},
					{`"uv_index_max":[4]`, `"uv_index_max":[21]`},
					{`"sunrise":[1793171520]`, `"sunrise":[true]`},
				} {
					var invalidRoot any
					invalid := json.NewDecoder(strings.NewReader(strings.Replace(tt.payload, mutation.old, mutation.replacement, 1)))
					invalid.UseNumber()
					if e := invalid.Decode(&invalidRoot); e != nil {
						t.Fatal(e)
					}
					if _, e := registry.Normalize(compiled.Definition, invalidRoot, context); e == nil {
						t.Fatalf("accepted invalid daily signal %q", mutation.replacement)
					}
				}
			}
			if tt.name == "met_office" {
				day := f.Daily[0]
				closeTo := func(value *float64, want float64) bool { return value != nil && math.Abs(*value-want) < 1e-9 }
				if !closeTo(day.WindGust, 8*3.6) || day.WindDirection == nil || *day.WindDirection != 220 ||
					day.Humidity == nil || *day.Humidity != 75 || !closeTo(day.Pressure, 101300*0.01) ||
					day.Visibility == nil || *day.Visibility != 20000 || day.FeelsLikeHigh == nil || *day.FeelsLikeHigh != 14 ||
					day.FeelsLikeLow == nil || *day.FeelsLikeLow != 7 || day.NightCondition == nil || *day.NightCondition != "cloudy" ||
					day.RainProbability == nil || *day.RainProbability != 60 || day.SnowProbability == nil || *day.SnowProbability != 0 ||
					day.UVMax == nil || *day.UVMax != 3 || f.Hourly[0].UV == nil || *f.Hourly[0].UV != 2 {
					t.Fatalf("missing Met Office daily values: %+v", day)
				}
			}
			context.ResourceID = "resource"
			context.Previous = &result.Resource
			next, e := registry.Normalize(compiled.Definition, root, context)
			if e != nil || next.Resource.ResourceID != "resource" || next.Resource.Revision != 2 {
				t.Fatal("lost refresh identity", e)
			}
			if tt.name == "open_meteo" {
				root.(map[string]any)["hourly"].(map[string]any)["temperature_2m"] = []any{}
				if _, e = registry.Normalize(compiled.Definition, root, context); e == nil {
					t.Fatal("accepted mismatched columns")
				}
			}
		})
	}
}

func TestShippedWeatherConditionMappings(t *testing.T) {
	for _, tc := range []struct{ provider, code, want string }{
		{"met_office", "27", "snow"}, {"open_meteo", "56", "drizzle"},
		{"open_meteo", "66", "rain"}, {"open_meteo", "85", "snow"},
		{"open_meteo", "97", "thunderstorm"},
	} {
		t.Run(tc.provider+"/"+tc.code, func(t *testing.T) {
			source, err := defaults.Files.ReadFile("weather/" + tc.provider + ".yaml")
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
			selectors, err := decode(compiled.Definition.Mapping)
			if err != nil {
				t.Fatal(err)
			}
			if got := selectors.Conditions[tc.code]; got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
