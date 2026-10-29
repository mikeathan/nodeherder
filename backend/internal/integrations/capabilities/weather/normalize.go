package weather

// Builds a validated forecast from provider data and location-specific time information.

import (
	"time"

	"node-herder/internal/integrations/capabilities"
	model "node-herder/models/integrations"
)

// forecastReader shares immutable selectors and timezone across one normalization.
type forecastReader struct {
	root      any
	selectors mapping
	zone      *time.Location
}

// newForecastReader resolves the provider or default IANA timezone before reading forecasts.
func newForecastReader(root any, selectors mapping) (forecastReader, error) {
	zoneName := text(pointer(root, selectors.Timezone))
	if zoneName == "" {
		zoneName = selectors.DefaultTimezone
	}
	if zoneName == "" {
		return forecastReader{}, invalidForecast("The provider omitted a timezone.")
	}
	zone, err := time.LoadLocation(zoneName)
	if err != nil {
		return forecastReader{}, invalidForecast("The provider returned an invalid timezone.")
	}
	return forecastReader{root: root, selectors: selectors, zone: zone}, nil
}

// forecast assembles a complete resource, preserving its identity and advancing its revision.
func (reader forecastReader) forecast(context capabilities.Context) (model.WeatherForecast, error) {
	forecast := model.WeatherForecast{
		ContractVersion: model.WeatherCapability,
		InstanceID:      context.InstanceID,
		ResourceID:      context.ResourceID,
		Revision:        1,
		Location:        text(pointer(reader.root, reader.selectors.Location)),
		Timezone:        reader.zone.String(),
		Attribution:     reader.selectors.Attribution,
		CapturedAt:      context.Now.UTC(),
		LastCheckedAt:   context.Now.UTC(),
	}
	if forecast.ResourceID == "" {
		id, err := context.NewIdentity("resource")
		if err != nil {
			return forecast, err
		}
		forecast.ResourceID = id
	}
	if forecast.Location == "" {
		forecast.Location = context.SelectedLabel
	}
	if forecast.Location == "" {
		forecast.Location = context.SelectedValue
	}
	// Refreshes retain the resource identity but advance its content revision.
	if context.Previous != nil {
		forecast.Revision = context.Previous.Revision + 1
	}
	var err error
	forecast.Hourly, err = reader.points(reader.selectors.Hourly)
	if err != nil {
		return forecast, err
	}
	forecast.Daily, err = reader.days()
	if err != nil {
		return forecast, err
	}
	forecast.Current, err = reader.current(forecast.Hourly, context.Now)
	if err != nil {
		return forecast, err
	}
	return forecast, forecast.Validate()
}

// points reads timestamped observations and converts their timestamps to UTC.
func (reader forecastReader) points(selectors series) ([]model.WeatherPoint, error) {
	sourceRows, err := rows(reader.root, selectors)
	if err != nil {
		return nil, invalidForecast("The provider returned invalid weather rows.")
	}
	points := make([]model.WeatherPoint, 0, len(sourceRows))
	for _, row := range sourceRows {
		timestamp, err := instant(row["time"], selectors.Format, reader.zone)
		if err != nil {
			return nil, invalidForecast("The provider returned an invalid forecast timestamp.")
		}
		points = append(points, model.WeatherPoint{Time: timestamp.UTC(), Condition: condition(row, reader.selectors), WeatherValues: values(row, selectors)})
	}
	return points, nil
}

// days reads daily summaries while preserving provider calendar dates.
func (reader forecastReader) days() ([]model.WeatherDay, error) {
	selectors := reader.selectors.Daily
	sourceRows, err := rows(reader.root, selectors)
	if err != nil {
		return nil, invalidForecast("The provider returned invalid daily weather.")
	}
	days := make([]model.WeatherDay, 0, len(sourceRows))
	for _, row := range sourceRows {
		date, err := reader.dailyDate(row["time"])
		if err != nil {
			return nil, err
		}
		days = append(days, model.WeatherDay{
			Date:                     date,
			Condition:                condition(row, reader.selectors),
			High:                     quantity(row, selectors, "high"),
			Low:                      quantity(row, selectors, "low"),
			Wind:                     quantity(row, selectors, "wind"),
			PrecipitationProbability: quantity(row, selectors, "precipitation_probability"),
			Precipitation:            quantity(row, selectors, "precipitation"),
		})
	}
	return days, nil
}

// dailyDate resolves a daily timestamp using the declared timezone or fixed provider offset.
func (reader forecastReader) dailyDate(value any) (string, error) {
	selectors := reader.selectors.Daily
	if selectors.Format == "date" {
		return text(value), nil
	}
	timestamp, err := instant(value, selectors.Format, reader.zone)
	if err != nil {
		return "", invalidForecast("The provider returned an invalid daily timestamp.")
	}
	if selectors.Offset == "" {
		return timestamp.In(reader.zone).Format("2006-01-02"), nil
	}
	offset := number(pointer(reader.root, selectors.Offset))
	if offset == nil {
		return "", invalidForecast("The provider omitted its daily timezone offset.")
	}
	if *offset < -50400 || *offset > 50400 {
		return "", invalidForecast("The provider returned an invalid daily timezone offset.")
	}
	// Some providers use the request-time offset for every daily boundary, including
	// dates crossing DST. Preserve that declared calendar instead of applying a new offset.
	return timestamp.UTC().Add(time.Duration(*offset) * time.Second).Format("2006-01-02"), nil
}

// current uses an explicit observation when supplied, otherwise the latest recent hourly point.
func (reader forecastReader) current(hourly []model.WeatherPoint, now time.Time) (*model.WeatherPoint, error) {
	if reader.selectors.Current != nil {
		points, err := reader.points(*reader.selectors.Current)
		if err != nil {
			return nil, err
		}
		if len(points) != 1 {
			return nil, invalidForecast("The provider returned invalid current weather.")
		}
		return &points[0], nil
	}
	// Do not present an old hourly entry as current weather after its observation window.
	var current *model.WeatherPoint
	for _, point := range hourly {
		if !point.Time.Before(now.Add(-time.Hour)) && !point.Time.After(now) {
			owned := point
			current = &owned
		}
	}
	return current, nil
}

func invalidForecast(message string) error { return model.Fail("invalid_response", message) }
