import { IconProps } from '@/types/icon.type';
import { KeyValuePair } from '../../types/types';

import {
  mdiWaterPercent,
  mdiWhiteBalanceSunny,
  mdiCloudDownload,
  mdiAtomVariant,
  mdiBolt,
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
  mdiDoorOpen,
  mdiThermometer,
  mdiThermometerHigh,
  mdiThermometerLow,
} from '@mdi/js';

// const typeToClassMapsensor: KeyValuePair<string> = {
//   humidity: 'text-info fa-tint',
//   illuminance: 'fa-sun',
//   pressure: 'fa-cloud-download-alt',
//   co2: 'fa-atom text-warning',
//   voltage: 'fa-bolt text-success',
//   state: 'fa-star-half-alt',
//   brightness: 'fa-sun',
//   occupancy: 'text-warning fa-walking',
//   current: 'fa-copyright text-warning',
//   power: 'fa-power-off text-success',
//   energy: 'fa-plug text-info',
//   frequency: 'fa-wave-square',
//   tamper: 'text-warning fa-exclamation-circle text-danger',
//   smoke: 'fa-smoking text-danger',
//   radiation_dose_per_hour: 'fa-radiation text-danger',
//   radioactive_events_per_minute:
//     'fa-radiation-alt text-warning',
//   power_factor: 'fa-industry text-danger',
//   mode: 'fa-user-cog text-warning',
//   sound: 'fa-volume-up text-info',
//   position: 'fa-percent text-info',
//   alarm: 'fa-exclamation-triangle text-danger',
//   color_xy: 'fa-palette',
//   color_hs: 'fa-palette',
//   color_temp: 'fa-sliders-h',
//   illuminance_lux: 'fa-sun',
//   soil_moisture: 'fa-fill-drip',
//   water_leak: 'fa-beat-fade text-primary fa-water',
//   week: 'fa-calendar-week',
//   workdays_schedule: 'fa-calendar-day text-info',
//   holidays_schedule: 'fa-calendar-day text-danger',
//   away_mode: 'fa-plane text-info',
//   vibration:
//     'fa-shake fa-rotate-270 text-primary fa-water fa-rotate-270',
//   power_outage_count: 'fa-plug-circle-xmark',
//   angle_x: 'fa-x',
//   angle_y: 'fa-y',
//   angle_z: 'fa-z',
//   side: 'fa-cube',
//   presence: 'fa-light fa-person',
//   contact: 'fa-fw fa-door-open',
// };

// TODO
// npm install @mdi/js
// <script setup>
// import { mdiSineWave } from '@mdi/js';
// </script>
// <template>
//   <component :is="mdiSineWave" :style="{ fontSize: '24px', color: 'blue' }"></component>
// </template>

// npm install @mdi/vue
// import { createApp } from 'vue';
// import App from './App.vue';
// import { Mdi } from '@mdi/vue';

// const app = createApp(App);
// app.component('Mdi', Mdi); // Register globally
// app.mount('#app');
// <template>
//   <Mdi icon="mdi-sine-wave" :size="36" color="green" />
// </template>

// const typeToClassMapsensor: KeyValuePair<string> = {
//   humidity: mdiWaterPercent,
//   illuminance: mdiWhiteBalanceSunny,
//   pressure: mdiCloudDownload,
//   co2: mdiAtomVariant,
//   voltage: mdiBolt,
//   state: mdiStarHalfFull,
//   brightness: mdiWhiteBalanceSunny,
//   occupancy: mdiWalk,
//   current: mdiCopyright,
//   power: mdiPower,
//   energy: mdiPowerPlug,
//   frequency: mdiSineWave,
//   tamper: mdiAlertCircleOutline,
//   smoke: mdiSmoking,
//   radiation_dose_per_hour: mdiRadioactive,
//   radioactive_events_per_minute: mdiRadioactive, // Consider a different icon if needed
//   power_factor: mdiFactory,
//   mode: mdiCog,
//   sound: mdiVolumeHigh,
//   position: mdiPercent,
//   alarm: mdiAlert,
//   color_xy: mdiPalette,
//   color_hs: mdiPalette,
//   color_temp: mdiThermometer, // Or a slider icon if more appropriate
//   illuminance_lux: mdiWhiteBalanceSunny,
//   soil_moisture: mdiWater, // Or a "water drop" icon
//   water_leak: mdiWater,
//   week: mdiCalendarWeek,
//   workdays_schedule: mdiCalendarCheck,
//   holidays_schedule: mdiCalendarCheck,
//   away_mode: mdiAirplane,
//   vibration: mdiVibrate,
//   power_outage_count: mdiPowerPlugOff,
//   angle_x: mdiAxisXArrow,
//   angle_y: mdiAxisYArrow,
//   angle_z: mdiAxisZArrow,
//   side: mdiCubeOutline,
//   presence: mdiAccount,
//   contact: mdiDoorOpen,
// };

const typeToClassMapsensor: KeyValuePair<IconProps> = {
  humidity: { name: mdiWaterPercent, color: 'blue' },
  illuminance: {
    name: mdiWhiteBalanceSunny,
    color: 'yellow',
  },
  pressure: { name: mdiCloudDownload, color: 'gray' },
  co2: { name: mdiAtomVariant, color: 'gray' },
  voltage: { name: mdiBolt, color: 'orange' },
  state: { name: mdiStarHalfFull, color: 'gold' },
  brightness: {
    name: mdiWhiteBalanceSunny,
    color: 'yellow',
  },
  occupancy: { name: mdiWalk, color: 'green' },
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
  color_temp: { name: mdiThermometer, color: 'orange' },
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
  presence: { name: mdiAccount, color: 'green' },
  contact: { name: mdiDoorOpen, color: 'gray' },
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
  value: number
): IconProps {
  switch (sensor) {
    case 'device_temperature':
    case 'temperature':
    case 'local_temperature':
      return getTemperatureIcon(value);
      break;
  }

  return typeToClassMapsensor[sensor];
}

const getTemperatureIcon = (
  temperature: number
): IconProps => {
  if (temperature >= 30) {
    return { name: mdiThermometerHigh, color: 'red' }; // High temperature
  } else if (temperature <= 10) {
    return { name: mdiThermometerLow, color: 'blue' }; // Low temperature
  } else {
    return { name: mdiThermometer, color: 'white' }; // Normal temperature (no specific color)
  }
};
