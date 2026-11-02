package integrations

import (
	"math"
	"time"
)

// WeatherCapability is the normalized, provider-independent weather resource contract.
const WeatherCapability = "weather_forecast/v1"

// WeatherIntegrationType is the catalogue category shared by weather providers.
const WeatherIntegrationType = "weather"

// Read-time freshness values.
const (
	WeatherFreshnessFresh = "fresh"
	WeatherFreshnessStale = "stale"
)

// Provider-independent weather condition vocabulary.
const (
	WeatherConditionUnknown      = "unknown"
	WeatherConditionClear        = "clear"
	WeatherConditionPartlyCloudy = "partly_cloudy"
	WeatherConditionCloudy       = "cloudy"
	WeatherConditionFog          = "fog"
	WeatherConditionDrizzle      = "drizzle"
	WeatherConditionRain         = "rain"
	WeatherConditionSnow         = "snow"
	WeatherConditionShowers      = "showers"
	WeatherConditionThunderstorm = "thunderstorm"
)

// Weather contract formats and bounded list/field sizes.
const (
	// WeatherClockLayout is the local HH:MM clock format in the resource timezone.
	WeatherClockLayout = "15:04"
	// WeatherDateLayout is the calendar-date format for daily entries.
	WeatherDateLayout = "2006-01-02"
	// WeatherMaxHourly bounds the hourly forecast horizon.
	WeatherMaxHourly = 384
	// WeatherMaxDaily bounds the daily forecast horizon.
	WeatherMaxDaily = 16
	// WeatherMaxLocationLength bounds the displayed location label.
	WeatherMaxLocationLength = 256
	// WeatherMaxAttributionLength bounds the provider attribution.
	WeatherMaxAttributionLength = 512
)

// Numeric validation bounds for normalized weather measurements.
const (
	weatherTemperatureMin   = -150
	weatherTemperatureMax   = 100
	weatherWindMax          = 600
	weatherDirectionMax     = 360
	weatherPercentMax       = 100
	weatherPrecipitationMax = 10000
	weatherPressureMin      = 800
	weatherPressureMax      = 1100
	weatherVisibilityMax    = 100000
	weatherUVMax            = 20
)

// WeatherValues contains optional quantities in Celsius, km/h, degrees, percent and mm.
// UV is the unitless UV index reported by the provider. Nil means unavailable; it must
// never become measured zero in presentation.
type WeatherValues struct {
	Temperature              *float64 `json:"temperature"`
	FeelsLike                *float64 `json:"feels_like"`
	Wind                     *float64 `json:"wind"`
	Gust                     *float64 `json:"gust"`
	Direction                *float64 `json:"direction"`
	Humidity                 *float64 `json:"humidity"`
	PrecipitationProbability *float64 `json:"precipitation_probability"`
	Precipitation            *float64 `json:"precipitation"`
	UV                       *float64 `json:"uv"`
}

// WeatherPoint is a forecast valid at a UTC instant, displayed in the resource timezone.
type WeatherPoint struct {
	Time      time.Time `json:"time"`
	Condition string    `json:"condition"`
	WeatherValues
}

// WeatherDay is a provider daily summary for a local calendar date; optional values retain gaps.
// Sunrise and Sunset are local HH:MM clock strings in the resource timezone. Pressure is hPa,
// Visibility is metres, wind values are km/h and probabilities are percent. Each field is nil
// when the provider does not supply it. NightCondition is nil when the provider has no separate
// night summary and "unknown" when its code is not mapped.
type WeatherDay struct {
	Date                     string   `json:"date"`
	Condition                string   `json:"condition"`
	NightCondition           *string  `json:"night_condition"`
	High                     *float64 `json:"high"`
	Low                      *float64 `json:"low"`
	FeelsLikeHigh            *float64 `json:"feels_like_high"`
	FeelsLikeLow             *float64 `json:"feels_like_low"`
	Wind                     *float64 `json:"wind"`
	WindGust                 *float64 `json:"wind_gust"`
	WindDirection            *float64 `json:"wind_direction"`
	Humidity                 *float64 `json:"humidity"`
	Pressure                 *float64 `json:"pressure"`
	Visibility               *float64 `json:"visibility"`
	PrecipitationProbability *float64 `json:"precipitation_probability"`
	Precipitation            *float64 `json:"precipitation"`
	RainProbability          *float64 `json:"rain_probability"`
	SnowProbability          *float64 `json:"snow_probability"`
	Sunrise                  *string  `json:"sunrise"`
	Sunset                   *string  `json:"sunset"`
	UVMax                    *float64 `json:"uv_max"`
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
	if len(f.Hourly) > WeatherMaxHourly || len(f.Daily) > WeatherMaxDaily {
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
	if len(forecast.Location) > WeatherMaxLocationLength || len(forecast.Attribution) > WeatherMaxAttributionLength {
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
	parsed, err := time.Parse(WeatherDateLayout, day.Date)
	if err != nil || parsed.Format(WeatherDateLayout) != day.Date {
		return false
	}
	if !validWeatherCondition(day.Condition) {
		return false
	}
	if day.High != nil && day.Low != nil && *day.Low > *day.High {
		return false
	}
	if !validWeatherClock(day.Sunrise) || !validWeatherClock(day.Sunset) {
		return false
	}
	if day.Sunrise != nil && day.Sunset != nil && *day.Sunset <= *day.Sunrise {
		return false
	}
	if day.NightCondition != nil && !validWeatherCondition(*day.NightCondition) {
		return false
	}
	return validMeasurements([]weatherMeasurement{
		{day.High, weatherTemperatureMin, weatherTemperatureMax},
		{day.Low, weatherTemperatureMin, weatherTemperatureMax},
		{day.FeelsLikeHigh, weatherTemperatureMin, weatherTemperatureMax},
		{day.FeelsLikeLow, weatherTemperatureMin, weatherTemperatureMax},
		{day.Wind, 0, weatherWindMax},
		{day.WindGust, 0, weatherWindMax},
		{day.WindDirection, 0, weatherDirectionMax},
		{day.Humidity, 0, weatherPercentMax},
		{day.Pressure, weatherPressureMin, weatherPressureMax},
		{day.Visibility, 0, weatherVisibilityMax},
		{day.PrecipitationProbability, 0, weatherPercentMax},
		{day.Precipitation, 0, weatherPrecipitationMax},
		{day.RainProbability, 0, weatherPercentMax},
		{day.SnowProbability, 0, weatherPercentMax},
		{day.UVMax, 0, weatherUVMax},
	})
}

// validWeatherClock accepts nil or a zero-padded local HH:MM string in the resource timezone.
func validWeatherClock(value *string) bool {
	if value == nil {
		return true
	}
	parsed, err := time.Parse(WeatherClockLayout, *value)
	return err == nil && parsed.Format(WeatherClockLayout) == *value
}

// Project evaluates age without modifying forecast contents, source identity or check time.
func (f WeatherForecast) Project(now time.Time) WeatherForecast {
	if f.Current != nil && (f.Current.Time.Before(now.Add(-time.Hour)) || f.Current.Time.After(now.Add(time.Hour))) {
		f.Current = nil
	}
	f.EvaluatedAt = now.UTC()
	f.Freshness = WeatherFreshnessFresh
	if !now.Before(f.CapturedAt.Add(time.Hour)) {
		f.Freshness = WeatherFreshnessStale
	}
	return f
}

func validWeatherPoint(point WeatherPoint) bool {
	if point.Time.IsZero() || !validWeatherCondition(point.Condition) {
		return false
	}
	return validMeasurements([]weatherMeasurement{
		{point.Temperature, weatherTemperatureMin, weatherTemperatureMax},
		{point.FeelsLike, weatherTemperatureMin, weatherTemperatureMax},
		{point.Wind, 0, weatherWindMax},
		{point.Gust, 0, weatherWindMax},
		{point.Direction, 0, weatherDirectionMax},
		{point.Humidity, 0, weatherPercentMax},
		{point.PrecipitationProbability, 0, weatherPercentMax},
		{point.Precipitation, 0, weatherPrecipitationMax},
		{point.UV, 0, weatherUVMax},
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
	case WeatherConditionUnknown, WeatherConditionClear, WeatherConditionPartlyCloudy,
		WeatherConditionCloudy, WeatherConditionFog, WeatherConditionDrizzle,
		WeatherConditionRain, WeatherConditionSnow, WeatherConditionShowers,
		WeatherConditionThunderstorm:
		return true
	}
	return false
}

func weatherRange(value *float64, min, max float64) bool {
	return value == nil || (!math.IsNaN(*value) && !math.IsInf(*value, 0) && *value >= min && *value <= max)
}
