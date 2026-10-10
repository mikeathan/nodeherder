/*
 * Icon registry (spec 007 design-system.md). Screens ask for icons by meaning ("home",
 * "delete"); the registry maps meanings to Material Design Icons paths. Expose and device
 * icons are resolved from generic metadata, so new devices get sensible icons without code
 * changes (FR-08). Colours come from kind-* CSS classes, never from here.
 */
import {
  mdiAccessPointNetwork, mdiAccountCircleOutline, mdiAirFilter, mdiAlarmLight, mdiAlertCircleOutline, mdiAlertOutline,
  mdiArrowDown, mdiArrowLeft, mdiArrowUp, mdiBattery, mdiBatteryAlertVariantOutline, mdiBrightness6, mdiCalendarClock,
  mdiCancel, mdiChartLine, mdiCheck, mdiChevronDown, mdiChevronLeft, mdiChevronRight, mdiChevronUp, mdiClockOutline,
  mdiClose, mdiCodeJson, mdiCogOutline, mdiConsoleLine, mdiContentCopy, mdiCreationOutline, mdiCubeOutline, mdiDevices,
  mdiDoorOpen, mdiDragVertical, mdiExportVariant, mdiFlashOutline, mdiFormatListBulleted, mdiGauge, mdiGestureTapButton,
  mdiGoogle, mdiHelpCircleOutline, mdiHomeOutline, mdiImport, mdiInformationOutline, mdiLanDisconnect, mdiLeaf,
  mdiLightbulbOnOutline, mdiLightbulbOutline, mdiLightningBolt, mdiLockOutline, mdiLogout, mdiMagnify, mdiMenu, mdiMinus,
  mdiMotionSensor, mdiPaletteOutline, mdiPencilOutline, mdiPlay, mdiPlus, mdiPowerPlugOutline, mdiRadar, mdiRefresh,
  mdiRemote, mdiRobotOutline, mdiShieldCheckOutline, mdiSignal, mdiSmokeDetectorVariant, mdiSync, mdiTabletDashboard,
  mdiThermometer, mdiTimerOutline, mdiTrashCanOutline, mdiTune, mdiViewDashboardOutline, mdiVolumeHigh, mdiWater,
  mdiWaterPercent, mdiWeatherNight, mdiWhiteBalanceSunny, mdiWifi, mdiZigbee,
} from '@mdi/js';
import { Device } from '@/types/device';
import { exposeKind } from '@/domain/exposes';

export const ICONS = {
  // navigation
  overview: mdiViewDashboardOutline,
  home: mdiHomeOutline,
  devices: mdiDevices,
  list: mdiFormatListBulleted,
  automation: mdiRobotOutline,
  assistant: mdiCreationOutline,
  console: mdiConsoleLine,
  settings: mdiCogOutline,
  join: mdiAccessPointNetwork,
  panel: mdiTabletDashboard,
  // actions
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

/** Icon that says what a device is, from what it exposes (no model lookups). */
export function deviceIcon(device: Device): string {
  const has = (name: string) => name in (device.exposes ?? {});
  if (has('brightness') || has('color_temp')) return mdiLightbulbOutline;
  if (has('state') && (has('power') || has('energy'))) return mdiPowerPlugOutline;
  if (has('state')) return mdiPowerPlugOutline;
  if (has('smoke') || has('alarm')) return mdiSmokeDetectorVariant;
  if (has('contact')) return mdiDoorOpen;
  if (has('presence') || has('occupancy')) return mdiMotionSensor;
  if (Object.keys(device.exposes ?? {}).some((n) => n.startsWith('action'))) return mdiRemote;
  if (has('temperature') || has('humidity')) return mdiThermometer;
  if (has('pm25') || has('voc') || has('co2')) return mdiAirFilter;
  return mdiCubeOutline;
}

export const protocolIcon = (connectionType: string): string =>
  connectionType === 'mqtt' ? mdiZigbee : connectionType === 'http' ? mdiWifi : mdiHelpCircleOutline;
