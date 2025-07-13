import { IconProps } from '@/types/icon.type';
import { KeyValuePair } from '../../types/types.type';

import {
  mdiWaterPercent,
  mdiWhiteBalanceSunny,
  mdiCloudDownload,
  mdiAtomVariant,
  mdiMoleculeCo,
  mdiMoleculeCo2,
  mdiLightningBolt,
  mdiStarHalfFull,
  mdiWalk,
  mdiCopyright,
  mdiPower,
  mdiPowerPlug,
  mdiSineWave,
  mdiAlertCircleOutline,
  mdiSmoking,
  mdiRadioactive,
  mdiFactory,
  mdiCog,
  mdiVolumeHigh,
  mdiPercent,
  mdiAlert,
  mdiPalette,
  mdiWater,
  mdiCalendarWeek,
  mdiCalendarCheck,
  mdiAirplane,
  mdiVibrate,
  mdiPowerPlugOff,
  mdiAxisXArrow,
  mdiAxisYArrow,
  mdiAxisZArrow,
  mdiCubeOutline,
  mdiAccount,
  mdiMotionSensorOff,
  mdiMotionSensor,
  mdiDoorOpen,
  mdiThermometer,
  mdiThermometerHigh,
  mdiThermometerLow,
  mdiTune,
  mdiVolumeVibrate,
  mdiLockOutline,
  mdiIntegratedCircuitChip,
  mdiBattery80,
  mdiBattery60,
  mdiBattery40,
  mdiBattery20,
  mdiBatteryAlert,
  mdiRadar,
  mdiAlarmLight,
  mdiAlarmLightOff,
  mdiCrosshairsQuestion,
  mdiBrightness6,
  mdiCreation,
  mdiPowerOff,
  mdiFlash,
} from '@mdi/js';
import { Expose } from '@/types/device';

const lampColor = '#ffc107';

const typeToClassMapsensor: KeyValuePair<IconProps> = {
  humidity: {
    name: mdiWaterPercent,
    color: 'rgba(13,202,240,1)',
  },
  device_fault: {
    name: mdiIntegratedCircuitChip,
    color: 'red',
  },
  illuminance: {
    name: mdiWhiteBalanceSunny,
    color: 'yellow',
  },
  pressure: { name: mdiCloudDownload, color: 'gray' },
  co2: { name: mdiMoleculeCo2, color: 'white' },
  co: { name: mdiMoleculeCo, color: 'white' },
  pm25: { name: mdiFactory, color: 'white' },
  voltage: { name: mdiLightningBolt, color: 'orange' },
  state: { name: mdiPower, color: 'white' },
  effect: { name: mdiCreation, color: 'gray' },

  brightness: {
    name: mdiBrightness6,
    color: lampColor,
  },
  occupancy: { name: mdiWalk, color: 'white' },
  current: { name: mdiCopyright, color: 'gray' },
  power: { name: mdiPower, color: 'red' },
  energy: { name: mdiFlash, color: 'green' },
  frequency: { name: mdiSineWave, color: 'purple' },
  tamper: { name: mdiAlertCircleOutline, color: 'red' },
  smoke: { name: mdiSmoking, color: 'red' },
  radiation_dose_per_hour: {
    name: mdiRadioactive,
    color: 'orange',
  },
  radioactive_events_per_minute: {
    name: mdiRadioactive,
    color: 'orange',
  },
  power_factor: { name: mdiFactory, color: 'gray' },
  mode: { name: mdiCog, color: 'gray' },
  sound: { name: mdiVolumeHigh, color: 'blue' },
  position: { name: mdiPercent, color: 'blue' },
  alarm: { name: mdiAlert, color: 'red' },
  color_xy: { name: mdiPalette, color: 'purple' },
  color_hs: { name: mdiPalette, color: 'purple' },
  color_temp: { name: mdiTune, color: 'white' },
  illuminance_lux: {
    name: mdiWhiteBalanceSunny,
    color: 'yellow',
  },
  soil_moisture: { name: mdiWater, color: 'brown' },
  water_leak: { name: mdiWater, color: 'blue' },
  week: { name: mdiCalendarWeek, color: 'gray' },
  workdays_schedule: {
    name: mdiCalendarCheck,
    color: 'green',
  },
  holidays_schedule: {
    name: mdiCalendarCheck,
    color: 'red',
  },
  away_mode: { name: mdiAirplane, color: 'blue' },
  vibration: { name: mdiVibrate, color: 'gray' },
  power_outage_count: {
    name: mdiPowerPlugOff,
    color: 'red',
  },
  target_distance: { name: mdiRadar, color: 'white' },
  angle_x: { name: mdiAxisXArrow, color: 'gray' },
  angle_y: { name: mdiAxisYArrow, color: 'gray' },
  angle_z: { name: mdiAxisZArrow, color: 'gray' },
  side: { name: mdiCubeOutline, color: 'gray' },
  presence: { name: mdiAccount, color: 'green' },
  contact: { name: mdiDoorOpen, color: 'white' },
  silence: { name: mdiVolumeVibrate, color: 'white' },
  child_lock: { name: mdiLockOutline, color: 'white' },
};

const sensorUnits: KeyValuePair<string> = {
  brightness: '%',
  temperature: '°C',
  pressure: 'hPa',
  humidity: '%',
  voltage: 'mV',
  linkquality: 'LQI',
  illuminance_lux: 'lux',
  illuminance: 'lux',
};

export function getFormattedSensorValue(expose: Expose): string {
  if (expose.name == null || expose.data == null) {
    return '-';
  }
  const unit = expose.unit ?? getSensorUnit(expose.name);
  switch (expose.name) {
    case 'presence':
      return expose.data ? 'Detected' : 'Clear';
    case 'contact':
      return expose.data ? 'Closed' : 'Open';
  }
  return `${getSensorValue(expose.data)}${unit}`;
}

export function getSensorUnit(sensor: string): string {
  if (sensor in sensorUnits === false) {
    return '';
  }
  return sensorUnits[sensor] as string;
}

export function getSensorName(sensor: string): string {
  if (!sensor) {
    return 'Sensor not found';
  }
  const formatted = sensor.replace('_', ' ');
  return `${formatted.charAt(0).toUpperCase()}${formatted.slice(1)}`;
}

export function getSensorValue(value: any): any {
  if (value == null) {
    return '—';
  }
  if (typeof value == 'boolean' || typeof value == 'string') {
    return value;
  }

  // todo : dont format integer values
  return parseFloat(value.toFixed(1));
}

export function getSensorIcon(sensor: string, value: any): IconProps {
  switch (sensor) {
    case 'alarm':
      return getAlarmIcon(value);
    case 'device_temperature':
    case 'temperature':
    case 'local_temperature':
      return getTemperatureIcon(value);
    case 'battery':
    case 'battpercentage':
      return getBatteryIcon(value);
    case 'state':
      return getStateIcon(value);
    case 'presence':
      return getPresenceIcon(value);
  }

  if (sensor?.startsWith('energy')) {
    return typeToClassMapsensor['energy'];
  }

  const icon = typeToClassMapsensor[sensor];
  return icon ?? getUnknownEntityIcon();
}

const getUnknownEntityIcon = (): IconProps => {
  return { name: mdiCrosshairsQuestion, color: 'grey' };
};

const getStateIcon = (value: boolean): IconProps => {
  return value ? { name: mdiPower, color: '#1E88E5' } : { name: mdiPower, color: 'grey' };
};

const getPresenceIcon = (value: boolean): IconProps => {
  return value ? { name: mdiMotionSensor, color: '#1E88E5' } : { name: mdiMotionSensorOff, color: 'grey' };
};

const getBatteryIcon = (value: number): IconProps => {
  if (value >= 80) {
    return { name: mdiBattery80, color: 'red' };
  } else if (value >= 60) {
    return { name: mdiBattery60, color: 'orange' };
  } else if (value >= 40) {
    return { name: mdiBattery40, color: 'yellow' };
  } else if (value >= 20) {
    return { name: mdiBattery20, color: 'yellow' };
  } else {
    return { name: mdiBatteryAlert, color: 'red' };
  }
};

const getAlarmIcon = (value: boolean): IconProps => {
  return value ? { name: mdiAlarmLight, color: 'red' } : { name: mdiAlarmLightOff, color: 'grey' };
};

const getTemperatureIcon = (temperature: number): IconProps => {
  if (temperature >= 30) {
    return { name: mdiThermometerHigh, color: 'red' }; // High temperature
  } else if (temperature <= 10) {
    return { name: mdiThermometerLow, color: '#1E88E5' }; // Low temperature
  } else {
    return {
      name: mdiThermometer,
      color: 'rgb(221, 111, 122)',
    }; // Normal temperature
  }
};
