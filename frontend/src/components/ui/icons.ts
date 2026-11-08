/*
 * Icon registry (spec 007 design-system.md). Screens ask for icons by meaning ("home",
 * "delete"); the registry maps meanings to Material Design Icons paths. Expose and device
 * icons are resolved from generic metadata, so new devices get sensible icons without code
 * changes (FR-08). Colours come from kind-* CSS classes, never from here.
 */
import {
  mdiAccessPointNetwork, mdiAccountCircleOutline, mdiAirFilter, mdiAlarmLight, mdiAlertCircleOutline, mdiAlertOutline,
  mdiApi, mdiArrowDown, mdiArrowLeft, mdiArrowUp, mdiBattery, mdiBatteryAlertVariantOutline, mdiBlur, mdiBrightness6, mdiCalendarClock,
  mdiCancel, mdiChartLine, mdiCheck, mdiChevronDown, mdiChevronLeft, mdiChevronRight, mdiChevronUp, mdiClockOutline,
  mdiClose, mdiCodeJson, mdiCogOutline, mdiCompassOutline, mdiConsoleLine, mdiContentCopy, mdiCreationOutline, mdiCubeOutline, mdiDevices,
  mdiDatabaseClock, mdiDoorOpen, mdiDotsVertical, mdiDragVertical, mdiExportVariant, mdiEyeOutline, mdiFlashOutline, mdiFlower, mdiFormatListBulleted, mdiGasStation, mdiGauge, mdiGestureTapButton,
  mdiGoogle, mdiHarddisk, mdiHelpCircleOutline, mdiHistory, mdiHomeOutline, mdiImport, mdiInformationOutline, mdiLanDisconnect, mdiLeaf,
  mdiLinkVariant, mdiLightbulbOnOutline, mdiLightbulbOutline, mdiLightningBolt, mdiLockOutline, mdiLogout, mdiMagnify, mdiMenu, mdiMinus,
  mdiMotionSensor, mdiPaletteOutline, mdiPencilOutline, mdiPlay, mdiPlus, mdiPowerPlugOutline, mdiRadar, mdiRefresh,
  mdiPin, mdiPinOutline, mdiPower, mdiRemote, mdiRobotOutline, mdiServerNetwork, mdiShieldCheckOutline, mdiSignal, mdiSmokeDetectorVariant, mdiSync, mdiTabletDashboard,
  mdiTextBoxOutline, mdiThermometer, mdiTimerOutline, mdiToggleSwitchVariant, mdiTrashCanOutline, mdiTune, mdiViewDashboardOutline, mdiVolumeHigh, mdiWater,
  mdiWaterPercent, mdiWeatherNight, mdiWeatherRainy, mdiWeatherSnowy, mdiWeatherSunny, mdiWeatherSunsetDown,
  mdiWeatherSunsetUp, mdiWeatherWindy, mdiWeatherWindyVariant, mdiWhiteBalanceSunny, mdiWifi, mdiZigbee,
} from '@mdi/js';
import { Device } from '@/types/device';
import { ConnectionTypes } from '@/types/device.type';
import { DERIVED_CONNECTION_TYPE } from '@/domain/devices';
import { exposeKind } from '@/domain/exposes';
import { DeviceIconKinds, deviceIconKind, type DeviceIconKind } from '@/domain/devices/icon';

export const ICONS = {
  // navigation
  overview: mdiViewDashboardOutline,
  home: mdiHomeOutline,
  devices: mdiDevices,
  presence: mdiMotionSensor,
  list: mdiFormatListBulleted,
  automation: mdiRobotOutline,
  assistant: mdiCreationOutline,
  console: mdiConsoleLine,
  settings: mdiCogOutline,
  join: mdiAccessPointNetwork,
  connections: mdiLinkVariant,
  panel: mdiTabletDashboard,
  // actions
  pin: mdiPin,
  pinOutline: mdiPinOutline,
  search: mdiMagnify,
  add: mdiPlus,
  remove: mdiMinus,
  delete: mdiTrashCanOutline,
  edit: mdiPencilOutline,
  run: mdiPlay,
  close: mdiClose,
  check: mdiCheck,
  back: mdiArrowLeft,
  menu: mdiMenu,
  more: mdiDotsVertical,
  tune: mdiTune,
  drag: mdiDragVertical,
  moveUp: mdiArrowUp,
  moveDown: mdiArrowDown,
  export: mdiExportVariant,
  import: mdiImport,
  copy: mdiContentCopy,
  json: mdiCodeJson,
  refresh: mdiRefresh,
  interview: mdiSync,
  signOut: mdiLogout,
  chevronLeft: mdiChevronLeft,
  chevronRight: mdiChevronRight,
  chevronDown: mdiChevronDown,
  chevronUp: mdiChevronUp,
  // states
  info: mdiInformationOutline,
  warn: mdiAlertOutline,
  error: mdiAlertCircleOutline,
  offline: mdiLanDisconnect,
  disabled: mdiCancel,
  unknown: mdiHelpCircleOutline,
  clock: mdiClockOutline,
  schedule: mdiCalendarClock,
  delay: mdiTimerOutline,
  chart: mdiChartLine,
  // appearance
  dark: mdiWeatherNight,
  light: mdiWhiteBalanceSunny,
  palette: mdiPaletteOutline,
  // settings sections (distinct from navigation and integration glyphs)
  settingsDevices: mdiToggleSwitchVariant,
  settingsStorage: mdiDatabaseClock,
  settingsUsage: mdiHarddisk,
  settingsLogger: mdiTextBoxOutline,
  settingsServer: mdiServerNetwork,
  settingsAssistant: mdiApi,
  settingsRetention: mdiHistory,
  // identity
  user: mdiAccountCircleOutline,
  google: mdiGoogle,
  secure: mdiShieldCheckOutline,
  zigbee: mdiZigbee,
  wifi: mdiWifi,
  battery: mdiBattery,
  batteryLow: mdiBatteryAlertVariantOutline,
  signal: mdiSignal,
  gauge: mdiGauge,
  power: mdiPower,
  // weather
  weatherTemperature: mdiThermometer,
  weatherRain: mdiWeatherRainy,
  weatherPrecipitation: mdiWater,
  weatherHumidity: mdiWaterPercent,
  weatherWind: mdiWeatherWindy,
  weatherGust: mdiWeatherWindyVariant,
  weatherDirection: mdiCompassOutline,
  weatherSun: mdiWeatherSunny,
  weatherPressure: mdiGauge,
  weatherVisibility: mdiEyeOutline,
  weatherSunrise: mdiWeatherSunsetUp,
  weatherSunset: mdiWeatherSunsetDown,
  weatherSnow: mdiWeatherSnowy,
  weatherNight: mdiWeatherNight,
  weatherAir: mdiAirFilter,
  weatherParticle: mdiBlur,
  weatherPollen: mdiFlower,
  fuel: mdiGasStation,
} as const;

export type IconName = keyof typeof ICONS;

const EXPOSE_ICONS: Record<string, string> = {
  temperature: mdiThermometer, local_temperature: mdiThermometer, device_temperature: mdiThermometer,
  humidity: mdiWaterPercent, soil_moisture: mdiWater, water_leak: mdiWater,
  brightness: mdiBrightness6, color_temp: mdiTune, illuminance: mdiWhiteBalanceSunny, illuminance_lux: mdiWhiteBalanceSunny,
  state: mdiLightbulbOutline,
  presence: mdiMotionSensor, occupancy: mdiMotionSensor, target_distance: mdiRadar,
  contact: mdiDoorOpen,
  smoke: mdiSmokeDetectorVariant, smoke_concentration: mdiSmokeDetectorVariant, alarm: mdiAlarmLight,
  power: mdiFlashOutline, voltage: mdiLightningBolt, current: mdiLightningBolt, energy: mdiLeaf,
  battery: mdiBattery, battpercentage: mdiBattery, battery_low: mdiBatteryAlertVariantOutline,
  linkquality: mdiSignal, rssi: mdiWifi,
  child_lock: mdiLockOutline, volume: mdiVolumeHigh, melody: mdiVolumeHigh,
  action: mdiGestureTapButton,
};

const KIND_ICONS: Record<string, string> = {
  temp: mdiThermometer, humidity: mdiWaterPercent, energy: mdiLeaf, power: mdiFlashOutline, light: mdiBrightness6,
  switch: mdiPowerPlugOutline, motion: mdiMotionSensor, contact: mdiDoorOpen, alarm: mdiAlarmLight, air: mdiAirFilter,
  battery: mdiBattery, signal: mdiSignal,
};

/** Icon for an expose by name, with a lit bulb for a light that is on. */
export function exposeIcon(name: string, on = false): string {
  if (name === 'state') return on ? mdiLightbulbOnOutline : mdiLightbulbOutline;
  if (name.startsWith('action')) return mdiGestureTapButton;
  return EXPOSE_ICONS[name] ?? KIND_ICONS[exposeKind(name)] ?? mdiCubeOutline;
}

/** Glyphs for the semantic device families resolved in the device domain. */
const DEVICE_ICONS: Readonly<Record<DeviceIconKind, string>> = {
  [DeviceIconKinds.Motion]: mdiMotionSensor,
  [DeviceIconKinds.Smoke]: mdiSmokeDetectorVariant,
  [DeviceIconKinds.Contact]: mdiDoorOpen,
  [DeviceIconKinds.Light]: mdiLightbulbOutline,
  [DeviceIconKinds.Power]: mdiPowerPlugOutline,
  [DeviceIconKinds.Remote]: mdiRemote,
  [DeviceIconKinds.Climate]: mdiThermometer,
  [DeviceIconKinds.Air]: mdiAirFilter,
  [DeviceIconKinds.Generic]: mdiCubeOutline,
};

/** Chooses a device glyph from discovered capabilities; the semantic choice lives in the domain. */
export function deviceIcon(device: Device): string {
  return DEVICE_ICONS[deviceIconKind(device)];
}

export const protocolIcon = (connectionType: string): string =>
  connectionType === ConnectionTypes.Mqtt
    ? mdiZigbee
    : connectionType === ConnectionTypes.Http || connectionType === ConnectionTypes.ESPHome
      ? mdiWifi
      : connectionType === DERIVED_CONNECTION_TYPE
        ? mdiRadar
        : mdiHelpCircleOutline;

/** Decorative symbols for normalized weather conditions; absent or unknown data stays explicit. */
export function weatherIcon(condition: string | undefined): string {
  return WEATHER_ICONS[condition ?? ''] ?? '—';
}

const WEATHER_ICONS: Readonly<Record<string, string>> = {
  clear: '☀️',
  partly_cloudy: '⛅',
  cloudy: '☁️',
  fog: '🌫️',
  drizzle: '🌦️',
  rain: '🌧️',
  snow: '🌨️',
  showers: '🌦️',
  thunderstorm: '⛈️',
};
