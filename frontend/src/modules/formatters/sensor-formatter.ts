import { KeyValuePair } from '../../types/types';

const typeToClassMapsensor: KeyValuePair<string> = {
  humidity: 'text-info fa-tint',
  illuminance: 'fa-sun',
  pressure: 'fa-cloud-download-alt',
  co2: 'fa-atom text-warning',
  voltage: 'fa-bolt text-success',
  state: 'fa-star-half-alt',
  brightness: 'fa-sun',
  occupancy: 'text-warning fa-walking',
  current: 'fa-copyright text-warning',
  power: 'fa-power-off text-success',
  energy: 'fa-plug text-info',
  frequency: 'fa-wave-square',
  tamper: 'text-warning fa-exclamation-circle text-danger',
  smoke: 'fa-smoking text-danger',
  radiation_dose_per_hour: 'fa-radiation text-danger',
  radioactive_events_per_minute:
    'fa-radiation-alt text-warning',
  power_factor: 'fa-industry text-danger',
  mode: 'fa-user-cog text-warning',
  sound: 'fa-volume-up text-info',
  position: 'fa-percent text-info',
  alarm: 'fa-exclamation-triangle text-danger',
  color_xy: 'fa-palette',
  color_hs: 'fa-palette',
  color_temp: 'fa-sliders-h',
  illuminance_lux: 'fa-sun',
  soil_moisture: 'fa-fill-drip',
  water_leak: 'fa-beat-fade text-primary fa-water',
  week: 'fa-calendar-week',
  workdays_schedule: 'fa-calendar-day text-info',
  holidays_schedule: 'fa-calendar-day text-danger',
  away_mode: 'fa-plane text-info',
  vibration:
    'fa-shake fa-rotate-270 text-primary fa-water fa-rotate-270',
  power_outage_count: 'fa-plug-circle-xmark',
  angle_x: 'fa-x',
  angle_y: 'fa-y',
  angle_z: 'fa-z',
  side: 'fa-cube',
  presence: 'fa-light fa-person', // <i class="fa-solid fa-person"></i>
  contact: 'fa-fw fa-door-open',
};

const sensorUnits: KeyValuePair<string> = {
  temperature: '°C',
  pressure: 'hPa',
  humidity: '%',
  voltage: 'mV',
  linkquality: 'LQI',
  illuminance_lux: 'lux',
};

export function getSensorUnit(sensor: string): string {
  if (sensor in sensorUnits === false) {
    return '';
  }
  return sensorUnits[sensor] as string;
}

export function getSensorName(sensor: string): string {
  return sensor.charAt(0).toUpperCase() + sensor.slice(1); // TODO: load name  from resources file
}

export function getSensorValue(value: any): any {
  if (value == null) {
    return '';
  }
  if (
    typeof value == 'boolean' ||
    typeof value == 'string'
  ) {
    return value;
  }

  // todo : dont format integer values
  return parseFloat(value.toFixed(1));
}

export function getSensorIcon(
  sensor: string,
  value: number,
): string {
  switch (sensor) {
    case 'device_temperature':
    case 'temperature':
    case 'local_temperature':
      typeToClassMapsensor[sensor] =
        getTemperatureIcon(value);
      break;
  }

  return typeToClassMapsensor[sensor];
}

const getTemperatureIcon = (
  temperature: number,
): string => {
  let icon = 'fa-thermometer-empty';
  if (temperature >= 30) {
    icon = 'fa-thermometer-full';
  } else if (temperature >= 25) {
    icon = 'fa-thermometer-three-quarters';
  } else if (temperature >= 20) {
    icon = 'fa-thermometer-half';
  } else if (temperature >= 15) {
    icon = 'fa-thermometer-quarter';
  }
  return 'text-danger ' + icon;
};
