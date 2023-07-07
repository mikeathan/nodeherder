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
  pressure: "hPa",
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
  if (typeof value == "boolean") {
    return value;
  }
  // todo : dont format integer values
  // todo;
  // format value with space and unit
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
  // user "text-danger".join(icon);
  return "text-danger " + icon;
};

// case 'contact':
//   classes.push(cx({ 'fa-door-closed text-muted': value, 'fa-door-open text-primary': !value }));
//   break;
// case 'occupancy':
//   classes.push(cx({ 'text-warning': value }));
//   break;
// case 'tamper':
//   classes.push(cx({ 'fa-beat-fade': value }));
//   break;
// case 'water_leak':
//   classes.push(cx({ 'fa-beat-fade text-primary': value }));
//   break;
// case 'vibration':
//   classes.push(cx({ 'fa-shake fa-rotate-270 text-primary': value }));
//   break;

// const typeToClassMap = {
//   humidity: ['text-info', 'fa-tint'],
//   illuminance: ['fa-sun'],
//   pressure: ['fa-cloud-download-alt'],
//   co2: ['fa-atom', 'text-warning'],
//   voltage: ['fa-bolt', 'text-success'],
//   state: ['fa-star-half-alt'],
//   brightness: ['fa-sun'],
//   occupancy: ['fa-walking'],
//   current: ['fa-copyright', 'text-warning'],
//   power: ['fa-power-off', 'text-success'],
//   energy: ['fa-plug', 'text-info'],
//   frequency: ['fa-wave-square'],
//   tamper: ['fa-exclamation-circle', 'text-danger'],
//   smoke: ['fa-smoking', 'text-danger'],
//   radiation_dose_per_hour: ['fa-radiation', 'text-danger'],
//   radioactive_events_per_minute: ['fa-radiation-alt', 'text-warning'],
//   power_factor: ['fa-industry', 'text-danger'],
//   mode: ['fa-user-cog', 'text-warning'],
//   sound: ['fa-volume-up', 'text-info'],
//   position: ['fa-percent', 'text-info'],
//   alarm: ['fa-exclamation-triangle', 'text-danger'],
//   color_xy: ['fa-palette'],
//   color_hs: ['fa-palette'],
//   color_temp: ['fa-sliders-h'],
//   illuminance_lux: ['fa-sun'],
//   soil_moisture: ['fa-fill-drip'],
//   water_leak: ['fa-water'],
//   week: ['fa-calendar-week'],
//   workdays_schedule: ['fa-calendar-day', 'text-info'],
//   holidays_schedule: ['fa-calendar-day', 'text-danger'],
//   away_mode: ['fa-plane', 'text-info'],
//   vibration: ['fa-water fa-rotate-270'],
//   power_outage_count: ['fa-plug-circle-xmark'],
//   angle_x: ['fa-x'],
//   angle_y: ['fa-y'],
//   angle_z: ['fa-z'],
//   side: ['fa-cube'],
// };
