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
  mdiCeilingLightMultiple
} from '@mdi/js';

const lampColor = '#ffc107';
const greyBackground ='#363636'
const typeToClassMapsensor: KeyValuePair<IconProps> = {
  humidity: {
    name: mdiWaterPercent,
    color: 'rgba(13,202,240,1)',
  },
  illuminance: {
    name: mdiWhiteBalanceSunny,
    color: lampColor,
  },
  pressure: { name: mdiCloudDownload, color: 'gray' },
  co2: { name: mdiMoleculeCo2, color: 'white' },
  co: { name: mdiMoleculeCo, color: 'white' },
  pm25: { name: mdiFactory, color: 'white' },
  voltage: { name: mdiLightningBolt, color: 'orange' },
  state: { name: mdiStarHalfFull, color: 'white' },
  brightness: {
    name: mdiCeilingLightMultiple,
    color: lampColor,
  },
  occupancy: { name: mdiWalk, color: 'white' },
  current: { name: mdiCopyright, color: 'gray' },
  power: { name: mdiPower, color: 'red' },
  energy: { name: mdiPowerPlug, color: 'green' },
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
  angle_x: { name: mdiAxisXArrow, color: 'gray' },
  angle_y: { name: mdiAxisYArrow, color: 'gray' },
  angle_z: { name: mdiAxisZArrow, color: 'gray' },
  side: { name: mdiCubeOutline, color: 'gray' },
  presence: { name: mdiAccount, color: 'black' },
  contact: { name: mdiDoorOpen, color: 'white' },
};

// const typeToClassMapsensor: KeyValuePair<string> = {
//   humidity: 'pi pi-droplet',
//   illuminance: 'pi pi-sun',
//   pressure: 'pi pi-cloud-download',
//   co2: 'mdi mdi-molecule-co2',
//   voltage: 'pi pi-bolt',
//   state: 'pi pi-star-half',
//   brightness: 'pi pi-sun',
//   occupancy: 'mdi mdi-walk',
//   current: 'mdi mdi-current-ac',
//   power: 'pi pi-power-off',
//   energy: 'mdi mdi-lightning-bolt',
//   frequency: 'mdi mdi-sine-wave',
//   tamper: 'pi pi-exclamation-triangle',
//   smoke: 'mdi mdi-smoke',
//   radiation_dose_per_hour: 'mdi mdi-radiation',
//   radioactive_events_per_minute: 'mdi mdi-radioactive',
//   power_factor: 'mdi mdi-chart-line',
//   mode: 'pi pi-cog',
//   sound: 'pi pi-volume-up',
//   position: 'mdi mdi-map-marker',
//   alarm: 'pi pi-bell',
//   color_xy: 'mdi mdi-palette',
//   color_hs: 'mdi mdi-palette',
//   color_temp: 'mdi mdi-thermometer',
//   illuminance_lux: 'pi pi-sun',
//   soil_moisture: 'mdi mdi-water-percent',
//   water_leak: 'mdi mdi-leak',
//   week: 'pi pi-calendar-clock',
//   workdays_schedule: 'pi pi-calendar',
//   holidays_schedule: 'pi pi-calendar-times',
//   away_mode: 'pi pi-send',
//   vibration: 'mdi mdi-vibrate',
//   power_outage_count: 'mdi mdi-power-plug-off',
//   angle_x: 'mdi mdi-axis-x-arrow',
//   angle_y: 'mdi mdi-axis-y-arrow',
//   angle_z: 'mdi mdi-axis-z-arrow',
//   side: 'mdi mdi-cube-outline',
//   presence: 'mdi mdi-account-check',
//   contact: 'mdi mdi-door', // mdi-door-closed, mdi-door-open
// };

const sensorUnits: KeyValuePair<string> = {
  temperature: '°C',
  pressure: 'hPa',
  humidity: '%',
  voltage: 'mV',
  linkquality: 'LQI',
  illuminance_lux: 'lux',
  illuminance: 'lux',
};



export function getEntityIcon(sensor: string, value: any): IconProps {

  switch (sensor) {
    case 'device_temperature':
    case 'temperature':
    case 'local_temperature':
      return getTemperatureIcon(value);
    case 'presence':
      return getPresenceIcon(value);
  }

  return typeToClassMapsensor[sensor];
}

const getPresenceIcon = (value: boolean): IconProps => {
  return value ? { name: mdiMotionSensor, color: '#1E88E5' } : { name: mdiMotionSensorOff, color: 'grey' };
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
