package weather

// Defines the typed weather field vocabulary shared by mapping, series and normalization.

// weatherField identifies one normalized weather value in provider mappings and rows.
type weatherField string

const (
	fieldTime                     weatherField = "time"
	fieldCondition                weatherField = "condition"
	fieldNightCondition           weatherField = "night_condition"
	fieldTemperature              weatherField = "temperature"
	fieldFeelsLike                weatherField = "feels_like"
	fieldWind                     weatherField = "wind"
	fieldGust                     weatherField = "gust"
	fieldDirection                weatherField = "direction"
	fieldHumidity                 weatherField = "humidity"
	fieldPressure                 weatherField = "pressure"
	fieldVisibility               weatherField = "visibility"
	fieldPrecipitationProbability weatherField = "precipitation_probability"
	fieldPrecipitation            weatherField = "precipitation"
	fieldUV                       weatherField = "uv"
	fieldHigh                     weatherField = "high"
	fieldLow                      weatherField = "low"
	fieldFeelsLikeHigh            weatherField = "feels_like_high"
	fieldFeelsLikeLow             weatherField = "feels_like_low"
	fieldWindGust                 weatherField = "wind_gust"
	fieldWindDirection            weatherField = "wind_direction"
	fieldRainProbability          weatherField = "rain_probability"
	fieldSnowProbability          weatherField = "snow_probability"
	fieldUVMax                    weatherField = "uv_max"
	fieldSunrise                  weatherField = "sunrise"
	fieldSunset                   weatherField = "sunset"
)
