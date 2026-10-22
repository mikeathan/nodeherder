/*
 * Expose capability and presentation registry (spec 007 FR-08).
 * Decides how any expose is shown and controlled from its generic metadata
 * (type, access mode, category, values, attributes), never from device models,
 * so new devices work without UI changes. Framework-free; covered by Jest.
 */
import { Device, Expose } from '@/types/device';
import { ExposeAccessModes, ExposeCategories, ExposeTypes } from '@/types/device.type';
import { getSensorName } from '@/modules/formatters/sensor-formatter';
import { displayNumber, exposeRange } from './controls';
export { exposeRange, type NumericRange } from './controls';

export type ExposeCapability =
  | 'switch'
  | 'slider'
  | 'preset'
  | 'binary-sensor'
  | 'event'
  | 'diagnostic'
  | 'config'
  | 'reading';

/** Colour family used for icons/tiles; each maps to a --nh-kind-* design token. */
export type ExposeKind =
  | 'temp'
  | 'humidity'
  | 'energy'
  | 'power'
  | 'light'
  | 'switch'
  | 'motion'
  | 'contact'
  | 'alarm'
  | 'air'
  | 'battery'
  | 'signal'
  | 'event'
  | 'misc';

const KIND_BY_NAME: Record<string, ExposeKind> = {
  temperature: 'temp',
  local_temperature: 'temp',
  device_temperature: 'temp',
  current_heating_setpoint: 'temp',
  humidity: 'humidity',
  soil_moisture: 'humidity',
  water_leak: 'humidity',
  energy: 'energy',
  energy_today: 'energy',
  energy_yesterday: 'energy',
  energy_month: 'energy',
  power: 'power',
  current: 'power',
  voltage: 'power',
  power_factor: 'power',
  brightness: 'light',
  color_temp: 'light',
  color: 'light',
  illuminance: 'light',
  illuminance_lux: 'light',
  state: 'switch',
  presence: 'motion',
  occupancy: 'motion',
  target_distance: 'motion',
  vibration: 'motion',
  contact: 'contact',
  alarm: 'alarm',
  smoke: 'alarm',
  smoke_concentration: 'alarm',
  device_fault: 'alarm',
  tamper: 'alarm',
  gas: 'alarm',
  water_leak_alarm: 'alarm',
  co2: 'air',
  voc: 'air',
  pm25: 'air',
  pm10: 'air',
  pm1: 'air',
  formaldehyd: 'air',
  pressure: 'air',
  air_quality_score: 'air',
  battery: 'battery',
  battpercentage: 'battery',
  battery_low: 'battery',
  linkquality: 'signal',
  rssi: 'signal',
};

const LABELS: Record<string, string> = {
  co2: 'CO₂',
  pm25: 'PM2.5',
  pm10: 'PM10',
  pm1: 'PM1',
  voc: 'VOC',
  formaldehyd: 'Formaldehyde',
  linkquality: 'Link quality',
  rssi: 'Wi-Fi signal',
  color_temp: 'Colour temperature',
  battpercentage: 'Battery',
  illuminance_lux: 'Illuminance',
  current_heating_setpoint: 'Setpoint',
  local_temperature: 'Temperature',
  energy_today: 'Energy today',
  energy_yesterday: 'Energy yesterday',
  energy_month: 'Energy this month',
};

/**
 * Words for binary values, in [on, off] order where "on" is values.on. Zigbee2MQTT defines
 * contact with on=false (open) and off=true (closed). Defaults to On/Off.
 */
const BINARY_WORDS: Record<string, [string, string]> = {
  contact: ['Open', 'Closed'],
  presence: ['Detected', 'Clear'],
  occupancy: ['Detected', 'Clear'],
  smoke: ['Smoke detected', 'Clear'],
  alarm: ['Sounding', 'Silent'],
  device_fault: ['Fault', 'OK'],
  tamper: ['Tampered', 'OK'],
  water_leak: ['Leak', 'Dry'],
  battery_low: ['Low', 'OK'],
  test: ['Testing', 'Idle'],
};

const humanise = (s: string) => s.replace(/_/g, ' ').replace(/^\w/, (c) => c.toUpperCase());

export const isWritable = (e: Expose): boolean =>
  [ExposeAccessModes.Readwrite, ExposeAccessModes.Write].includes(e.access_mode as 'readwrite' | 'write');

export function exposeKind(name: string): ExposeKind {
  if (KIND_BY_NAME[name]) return KIND_BY_NAME[name];
  if (name.startsWith('energy')) return 'energy';
  if (name.startsWith('action') || name === 'event_count') return 'event';
  return 'misc';
}

/** Shared presentation palette: timeline segments, legends and state badges must
 * agree on which value is open/closed or alert/clear. Unknown uses its own pattern. */
export function exposePalette(expose: Expose): { active: string; inactive: string } {
  if (expose.name === 'contact')
    return { active: 'var(--nh-warn)', inactive: 'var(--nh-kind-contact)' };
  if (exposeKind(expose.name) === 'alarm' || expose.name === 'water_leak' || expose.name === 'battery_low')
    return { active: 'var(--nh-danger)', inactive: 'var(--nh-kind-contact)' };
  return { active: `var(--nh-kind-${exposeKind(expose.name)})`, inactive: 'var(--nh-text-muted)' };
}

/** A current reading and its matching historical report use exactly the same colour. */
export function exposeValueColor(expose: Expose, value: unknown = expose.data): string {
  if (expose.type !== 'binary') return `var(--nh-kind-${exposeKind(expose.name)})`;
  const state = binaryStateOf(expose, value);
  if (state === null) return 'var(--nh-text-muted)';
  const palette = exposePalette(expose);
  return state ? palette.active : palette.inactive;
}

export function exposeLabel(name: string): string {
  return LABELS[name] ?? getSensorName(name).replace(/_/g, ' ');
}

/** Value of the "on" side for a binary expose; Zigbee2MQTT uses values.on/off, others plain booleans. */
export function binaryOnValue(e: Expose): unknown {
  return e.values && 'on' in e.values ? e.values['on'] : true;
}
export function binaryOffValue(e: Expose): unknown {
  return e.values && 'off' in e.values ? e.values['off'] : false;
}

/** A toggle is a device-declared command, never a confirmed binary state. */
function binaryToggleValue(e: Expose): string | undefined {
  const value = e.values?.['toggle'];
  return e.type === ExposeTypes.Binary && typeof value === 'string' && value.trim() !== '' &&
    value !== binaryOnValue(e) && value !== binaryOffValue(e) ? value : undefined;
}

export function binaryValueOptions(e: Expose, allowCommands = false): { value: unknown; label: string }[] {
  const options = [binaryOnValue(e), binaryOffValue(e)].map((value) => ({ value, label: formatExposeValue(e, value) }));
  const toggle = allowCommands ? binaryToggleValue(e) : undefined;
  if (toggle !== undefined) options.push({ value: toggle, label: 'Toggle' });
  return options;
}

/** Action presentation includes supported commands; readings/conditions stay strict. */
export function formatExposeCommandValue(e: Expose, value: unknown): string {
  const toggle = binaryToggleValue(e);
  return toggle !== undefined && value === toggle ? 'Toggle' : formatExposeValue(e, value);
}
export function isBinaryOn(e: Expose, value: unknown = e.data): boolean {
  return binaryStateOf(e, value) === true;
}

/** Confirmed binary state; `null` means the latest value is missing or outside values.on/off. */
export function binaryStateOf(e: Expose, value: unknown = e.data): boolean | null {
  if (value === binaryOnValue(e)) return true;
  if (value === binaryOffValue(e)) return false;
  // Raw history is typed as boolean while live z2m words are often "ON"/"OFF".
  if (typeof value === 'boolean' && typeof binaryOnValue(e) !== 'boolean') return value;
  return null;
}

/** Human warmth description for a Kelvin reading; no invented device capabilities. */
export function colourTemperatureName(kelvin: number): string {
  if (!Number.isFinite(kelvin) || kelvin <= 0) return 'Unknown';
  if (kelvin < 3000) return 'Warm white';
  if (kelvin < 4500) return 'Neutral white';
  return 'Cool white';
}

/** Named preset values (e.g. colour temperature "warm" → 454), excluding sentinel values. */
export function exposePresets(e: Expose): { label: string; value: unknown }[] {
  if (!e.values) return [];
  const range = exposeRange(e);
  return Object.entries(e.values)
    .filter(([, v]) => !(range && typeof v === 'number' && (v < range.min || v > range.max)))
    .map(([k, v]) =>
      e.type === ExposeTypes.Enum ? { label: humanise(String(v)), value: v } : { label: humanise(k), value: v }
    );
}

export function capabilityOf(e: Expose): ExposeCapability {
  if (e.category === ExposeCategories.Config) return 'config';
  const control = controlOf(e);
  if (control === 'switch' || control === 'slider') return control;
  if (isPresetExpose(e)) return 'preset';
  if (e.type === ExposeTypes.Binary) return 'binary-sensor';
  if (e.name.startsWith('action')) return 'event';
  return e.category === ExposeCategories.Diagnostic ? 'diagnostic' : 'reading';
}

function isPresetExpose(expose: Expose): boolean {
  const supported = [ExposeTypes.Numeric, ExposeTypes.Enum].includes(expose.type as 'numeric' | 'enum');
  return supported && isWritable(expose) && exposePresets(expose).length > 0;
}

export type ExposeControl = 'switch' | 'slider' | 'choice' | 'number' | 'none';

/** Which input changes this expose (any category; config exposes are edited the same way). */
export function controlOf(e: Expose): ExposeControl {
  if (!isWritable(e)) return 'none';
  if (e.type === ExposeTypes.Binary) return 'switch';
  if (e.type === ExposeTypes.Numeric) return exposeRange(e) ? 'slider' : 'number';
  if (e.type === ExposeTypes.Enum && exposePresets(e).length) return 'choice';
  return 'none';
}

const numberText = (v: number): string =>
  Number.isInteger(v) ? String(v) : String(parseFloat(v.toFixed(Math.abs(v) < 10 ? 2 : 1)));

/**
 * Display text for an expose value (defaults to its current data). Uses expose metadata when
 * present: brightness becomes a percentage of its maximum; binary values become words.
 * Missing data is "—", never a fake zero (NH-06).
 */
export function formatExposeValue(e: Expose, value: unknown = e.data): string {
  if (value === null || value === undefined || value === '') return '—';
  if (e.type === ExposeTypes.Binary) return binaryText(e, value);
  if (typeof value === 'number') return numericText(e, value);
  if (typeof value === 'string') return stringText(value);
  return '—';
}
function binaryText(e: Expose, value: unknown): string {
  const [on, off] = BINARY_WORDS[e.name] ?? ['On', 'Off'];
  const state = binaryStateOf(e, value);
  if (state === true) return on;
  if (state === false) return off;
  return 'Unknown';
}
function numericText(e: Expose, value: number): string {
  if (!Number.isFinite(value)) return '—';
  if (e.name === 'brightness') return `${Math.round(displayNumber(e, value))} %`;
  if (e.name === 'color_temp') return value > 0 ? `${Math.round(displayNumber(e, value))} K` : '—';
  const unit = e.unit && e.unit !== 'lqi' ? e.unit : '';
  return unit ? `${numberText(value)} ${unit}` : numberText(value);
}
function stringText(value: string): string {
  if (value === 'ON') return 'On';
  if (value === 'OFF') return 'Off';
  return humanise(value);
}

/** The device's switch expose (binary, writable, measurement), if any. */
export function stateExposeOf(device: Device): Expose | undefined {
  return Object.values(device.exposes).find(
    (e) => e.type === ExposeTypes.Binary && isWritable(e) && e.category === ExposeCategories.Measurement
  );
}

/** Exposes shown as the device's readings/controls (excludes diagnostics and config). */
export function measurementExposes(device: Device): Expose[] {
  return Object.values(device.exposes).filter((e) => e.category === ExposeCategories.Measurement);
}
