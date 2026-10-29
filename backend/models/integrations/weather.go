package integrations

import (
	"math"
	"time"
)

// WeatherCapability is the normalized, provider-independent weather resource contract.
const WeatherCapability = "weather_forecast/v1"

// WeatherValues contains optional quantities in Celsius, km/h, degrees, percent and mm.
// Nil means unavailable; it must never become measured zero in presentation.
type WeatherValues struct {
	Temperature              *float64 `json:"temperature"`
	FeelsLike                *float64 `json:"feels_like"`
	Wind                     *float64 `json:"wind"`
	Gust                     *float64 `json:"gust"`
	Direction                *float64 `json:"direction"`
	Humidity                 *float64 `json:"humidity"`
	PrecipitationProbability *float64 `json:"precipitation_probability"`
	Precipitation            *float64 `json:"precipitation"`
}

// WeatherPoint is a forecast valid at a UTC instant, displayed in the resource timezone.
type WeatherPoint struct {
	Time      time.Time `json:"time"`
	Condition string    `json:"condition"`
	WeatherValues
}

// WeatherDay is a provider daily summary for a local calendar date; optional values retain gaps.
type WeatherDay struct {
	Date                     string   `json:"date"`
	Condition                string   `json:"condition"`
	High                     *float64 `json:"high"`
	Low                      *float64 `json:"low"`
	Wind                     *float64 `json:"wind"`
	PrecipitationProbability *float64 `json:"precipitation_probability"`
	Precipitation            *float64 `json:"precipitation"`
}

// WeatherForecast is a bounded snapshot; refresh and storage remain owned by integration runtime.
// Hourly instants and daily dates are strictly ascending. Projection does not refresh providers.
type WeatherForecast struct {
	ContractVersion string         `json:"contract_version"`
	InstanceID      string         `json:"instance_id"`
	ResourceID      string         `json:"resource_id"`
	Revision        uint64         `json:"revision"`
	Location        string         `json:"location"`
	Timezone        string         `json:"timezone"`
	Attribution     string         `json:"attribution"`
	CapturedAt      time.Time      `json:"captured_at"`
	LastCheckedAt   time.Time      `json:"last_checked_at"`
	EvaluatedAt     time.Time      `json:"evaluated_at"`
	Freshness       string         `json:"freshness"`
	Current         *WeatherPoint  `json:"current"`
	Hourly          []WeatherPoint `json:"hourly"`
	Daily           []WeatherDay   `json:"daily"`
}

// Validate rejects invalid saved or normalized resources independently of provider decoding.
func (f WeatherForecast) Validate() error {
	invalid := Fail("invalid_resource", "The weather data is invalid. Refresh or check the provider configuration.")
	if !validForecastHeader(f) {
		return invalid
	}
	if len(f.Hourly) > 384 || len(f.Daily) > 16 {
		return invalid
	}
	if len(f.Hourly)+len(f.Daily) == 0 {
		return invalid
	}
	if f.Current != nil && !validWeatherPoint(*f.Current) {
		return invalid
	}
	if !validHourlyWeather(f.Hourly) {
		return invalid
	}
	if !validDailyWeather(f.Daily) {
		return invalid
	}
	return nil
}

func validForecastHeader(forecast WeatherForecast) bool {
	if forecast.ContractVersion != WeatherCapability || forecast.Revision == 0 {
		return false
	}
	for _, value := range []string{forecast.InstanceID, forecast.ResourceID, forecast.Location, forecast.Attribution} {
		if value == "" {
			return false
		}
	}
	if len(forecast.Location) > 256 || len(forecast.Attribution) > 512 {
		return false
	}
	if forecast.CapturedAt.IsZero() || forecast.LastCheckedAt.IsZero() {
		return false
	}
	if forecast.Timezone == "" || forecast.Timezone == "Local" {
		return false
	}
	_, err := time.LoadLocation(forecast.Timezone)
	return err == nil
}

func validHourlyWeather(points []WeatherPoint) bool {
	var previous time.Time
	for _, point := range points {
		if !validWeatherPoint(point) || !point.Time.After(previous) {
			return false
		}
		previous = point.Time
	}
	return true
}

func validDailyWeather(days []WeatherDay) bool {
	previous := ""
	for _, day := range days {
		if day.Date <= previous || !validWeatherDay(day) {
			return false
		}
		previous = day.Date
	}
	return true
}

func validWeatherDay(day WeatherDay) bool {
	parsed, err := time.Parse("2006-01-02", day.Date)
	if err != nil || parsed.Format("2006-01-02") != day.Date {
		return false
	}
	if !validWeatherCondition(day.Condition) {
		return false
	}
	if day.High != nil && day.Low != nil && *day.Low > *day.High {
		return false
	}
	return validMeasurements([]weatherMeasurement{{day.High, -150, 100}, {day.Low, -150, 100}, {day.Wind, 0, 600}, {day.PrecipitationProbability, 0, 100}, {day.Precipitation, 0, 10000}})
}

// Project evaluates age without modifying forecast contents, source identity or check time.
func (f WeatherForecast) Project(now time.Time) WeatherForecast {
	if f.Current != nil && (f.Current.Time.Before(now.Add(-time.Hour)) || f.Current.Time.After(now.Add(time.Hour))) {
		f.Current = nil
	}
	f.EvaluatedAt = now.UTC()
	f.Freshness = "fresh"
	if !now.Before(f.CapturedAt.Add(time.Hour)) {
		f.Freshness = "stale"
	}
	return f
}

func validWeatherPoint(point WeatherPoint) bool {
	if point.Time.IsZero() || !validWeatherCondition(point.Condition) {
		return false
	}
	return validMeasurements([]weatherMeasurement{
		{point.Temperature, -150, 100}, {point.FeelsLike, -150, 100},
		{point.Wind, 0, 600}, {point.Gust, 0, 600}, {point.Direction, 0, 360},
		{point.Humidity, 0, 100}, {point.PrecipitationProbability, 0, 100}, {point.Precipitation, 0, 10000},
	})
}

type weatherMeasurement struct {
	value    *float64
	min, max float64
}

func validMeasurements(measurements []weatherMeasurement) bool {
	for _, measurement := range measurements {
		if !weatherRange(measurement.value, measurement.min, measurement.max) {
			return false
		}
	}
	return true
}

func validWeatherCondition(value string) bool {
	switch value {
	case "unknown", "clear", "partly_cloudy", "cloudy", "fog", "drizzle", "rain", "snow", "showers", "thunderstorm":
		return true
	}
	return false
}

func weatherRange(value *float64, min, max float64) bool {
	return value == nil || (!math.IsNaN(*value) && !math.IsInf(*value, 0) && *value >= min && *value <= max)
}
