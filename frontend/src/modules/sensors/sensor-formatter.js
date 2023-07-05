const typeToClassMap = {
  temperature: "fa-thermometer-full",
  humidity: "fa-tint",
  illuminance: "fa-sun",
  pressure: "fa-cloud-download-alt",
  co2: "text-warning",
  voltage: "text-success",
  state: "fa-star-half-alt",
  brightness: "fa-sun",
  occupancy: "fa-walking",
  current: "fa-copyright",
  power: "fa-power-off",
  energy: "fa-plug",
  frequency: "fa-wave-square",
  tamper: "fa-exclamation-circle",
  smoke: "fa-smoking",
  radiation_dose_per_hour: "fa-radiation",
  radioactive_events_per_minute: "fa-radiation-alt",
  power_factor: "fa-industry",
  mode: "fa-user-cog",
  sound: "fa-volume-up",
  position: "fa-percent",
  alarm: "fa-exclamation-triangle",
  color_xy: "fa-palette",
  color_hs: "fa-palette",
  color_temp: "fa-sliders-h",
  illuminance_lux: "fa-sun",
  soil_moisture: "fa-fill-drip",
  water_leak: "fa-water",
  week: "fa-calendar-week",
  workdays_schedule: "fa-calendar-day",
  holidays_schedule: "fa-calendar-day",
  away_mode: "fa-plane",
};
const sensorUnits = {
  temperature: "°C",
  pressure: "$hPa",
  humidity: "%",
  voltage: "mV",
  linkquality: "LQI",
  illuminance_lux: "lux",
};

function getUnit(sensor) {
  if (sensorUnits[sensor] == undefined) {
    return "";
  }
  return sensorUnits[sensor];
}
export function getSensorName(sensor) {
  return sensor.charAt(0).toUpperCase() + sensor.slice(1); // TODO: load name  from resources file
}
export function getSensorValue(sensor, value) {
  return value.toFixed(1) + getUnit(sensor);
}

export function getSensorIcon(sensor, value) {
  switch (sensor) {
    case "device_temperature":
    case "temperature":
    case "local_temperature":
      typeToClassMap[sensor] = getTemperatureIcon(value);
      break;
  }

  return typeToClassMap[sensor];
}

const getTemperatureIcon = (temperature) => {
  let icon = "fa-thermometer-empty";
  if (temperature >= 30) {
    icon = "fa-thermometer-full";
  } else if (temperature >= 25) {
    icon = "fa-thermometer-three-quarters";
  } else if (temperature >= 20) {
    icon = "fa-thermometer-half";
  } else if (temperature >= 15) {
    icon = "fa-thermometer-quarter";
  }
  return icon;
};
