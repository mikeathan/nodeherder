package integrations

import (
	"math"
	"testing"
	"time"
)

func TestWeatherForecastValidation(t *testing.T) {
	now := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	fixture := func() WeatherForecast {
		return WeatherForecast{ContractVersion: WeatherCapability, InstanceID: "connection", ResourceID: "weather", Revision: 1, Location: "Bristol", Timezone: "Europe/London", Attribution: "Provider", CapturedAt: now, LastCheckedAt: now, Hourly: []WeatherPoint{{Time: now, Condition: "rain"}}}
	}
	tests := []struct {
		name   string
		change func(*WeatherForecast)
	}{
		{"invalid timezone", func(f *WeatherForecast) { f.Timezone = "invalid" }},
		{"implicit timezone", func(f *WeatherForecast) { f.Timezone = "Local" }},
		{"duplicate hours", func(f *WeatherForecast) { f.Hourly = append(f.Hourly, f.Hourly[0]) }},
		{"invalid probability", func(f *WeatherForecast) { v := 101.0; f.Hourly[0].PrecipitationProbability = &v }},
		{"invalid uv", func(f *WeatherForecast) { v := 21.0; f.Hourly[0].UV = &v }},
		{"negative daily uv", func(f *WeatherForecast) {
			v := -1.0
			f.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "clear", UVMax: &v}}
		}},
		{"invalid sunrise", func(f *WeatherForecast) {
			value := "7:05"
			f.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "clear", Sunrise: &value}}
		}},
		{"sunset before sunrise", func(f *WeatherForecast) {
			sunrise, sunset := "19:00", "06:00"
			f.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "clear", Sunrise: &sunrise, Sunset: &sunset}}
		}},
		{"invalid feels like", func(f *WeatherForecast) {
			v := 101.0
			f.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "clear", FeelsLikeHigh: &v}}
		}},
		{"invalid wind gust", func(f *WeatherForecast) {
			v := 601.0
			f.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "clear", WindGust: &v}}
		}},
		{"invalid wind direction", func(f *WeatherForecast) {
			v := 361.0
			f.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "clear", WindDirection: &v}}
		}},
		{"invalid humidity", func(f *WeatherForecast) {
			v := 101.0
			f.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "clear", Humidity: &v}}
		}},
		{"invalid pressure", func(f *WeatherForecast) {
			v := 1200.0
			f.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "clear", Pressure: &v}}
		}},
		{"invalid visibility", func(f *WeatherForecast) {
			v := -1.0
			f.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "clear", Visibility: &v}}
		}},
		{"invalid snow probability", func(f *WeatherForecast) {
			v := 101.0
			f.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "clear", SnowProbability: &v}}
		}},
		{"invalid rain probability", func(f *WeatherForecast) {
			v := -1.0
			f.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "clear", RainProbability: &v}}
		}},
		{"invalid night condition", func(f *WeatherForecast) {
			v := "stormy"
			f.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "clear", NightCondition: &v}}
		}},
		{"non finite", func(f *WeatherForecast) { v := math.NaN(); f.Hourly[0].Temperature = &v }},
		{"invalid date", func(f *WeatherForecast) { f.Daily = []WeatherDay{{Date: "2026-02-30", Condition: "unknown"}} }},
		{"inverted temperatures", func(f *WeatherForecast) {
			high, low := 1.0, 2.0
			f.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "unknown", High: &high, Low: &low}}
		}},
		{"empty", func(f *WeatherForecast) { f.Hourly = nil }},
	}
	if err := fixture().Validate(); err != nil {
		t.Fatal(err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := fixture()
			tt.change(&f)
			if f.Validate() == nil {
				t.Fatal("invalid forecast accepted")
			}
		})
	}
	withSignals := fixture()
	hourlyUV, dailyUV := 5.0, 6.0
	sunrise, sunset := "07:00", "19:00"
	withSignals.Hourly[0].UV = &hourlyUV
	withSignals.Daily = []WeatherDay{{Date: "2026-10-28", Condition: "clear", Sunrise: &sunrise, Sunset: &sunset, UVMax: &dailyUV}}
	if err := withSignals.Validate(); err != nil {
		t.Fatal("valid daily signals rejected", err)
	}
	if projected := withSignals.Project(now.Add(time.Hour)); projected.Daily[0].UVMax == nil || *projected.Daily[0].UVMax != dailyUV {
		t.Fatal("projection dropped daily signals")
	}
	f := fixture()
	projected := f.Project(now.Add(time.Hour))
	if projected.Freshness != "stale" || !f.EvaluatedAt.IsZero() || projected.Hourly[0].Temperature != nil {
		t.Fatal("projection changed evidence or invented data")
	}
	if f.Project(now.Add(59*time.Minute)).Freshness != "fresh" {
		t.Fatal("premature stale projection")
	}
	resource, err := NewResource(f)
	if err != nil {
		t.Fatal(err)
	}
	var decoded WeatherForecast
	if err = resource.Decode(&decoded); err != nil || decoded.Validate() != nil {
		t.Fatal("resource round trip failed", err)
	}
}
