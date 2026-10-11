/*
 * Expose capability and presentation registry (spec 007 FR-08).
 * Decides how any expose is shown and controlled from its generic metadata
 * (type, access mode, category, values, attributes), never from device models,
 * so new devices work without UI changes. Framework-free; covered by Jest.
 */
import { Device, Expose } from '@/types/device';
import { ExposeAccessModes, ExposeCategories, ExposeTypes } from '@/types/device.type';
import { getFormattedSensorValueByName, getSensorName } from '@/modules/formatters/sensor-formatter';

export type ExposeCapability = 'switch' | 'slider' | 'preset' | 'binary-sensor' | 'event' | 'diagnostic' | 'config' | 'reading';

/** Colour family used for icons/tiles; each maps to a --nh-kind-* design token. */
export type ExposeKind = 'temp' | 'humidity' | 'energy' | 'power' | 'light' | 'switch' | 'motion' | 'contact' | 'alarm' | 'air' | 'battery' | 'signal' | 'misc';

const KIND_BY_NAME: Record<string, ExposeKind> = {
  temperature: 'temp', local_temperature: 'temp', device_temperature: 'temp', current_heating_setpoint: 'temp',
  humidity: 'humidity', soil_moisture: 'humidity', water_leak: 'humidity',
  energy: 'energy', energy_today: 'energy', energy_yesterday: 'energy', energy_month: 'energy',
  power: 'power', current: 'power', voltage: 'power', power_factor: 'power',
  brightness: 'light', color_temp: 'light', color: 'light', illuminance: 'light', illuminance_lux: 'light',
  state: 'switch',
  presence: 'motion', occupancy: 'motion', target_distance: 'motion', vibration: 'motion',
  contact: 'contact',
  alarm: 'alarm', smoke: 'alarm', smoke_concentration: 'alarm', device_fault: 'alarm', tamper: 'alarm', gas: 'alarm', water_leak_alarm: 'alarm',
  co2: 'air', voc: 'air', pm25: 'air', pm10: 'air', pm1: 'air', formaldehyd: 'air', pressure: 'air', air_quality_score: 'air',
  battery: 'battery', battpercentage: 'battery', battery_low: 'battery',
  linkquality: 'signal', rssi: 'signal',
};

const LABELS: Record<string, string> = {
  co2: 'CO₂', pm25: 'PM2.5', pm10: 'PM10', pm1: 'PM1', voc: 'VOC', formaldehyd: 'Formaldehyde',
  linkquality: 'Link quality', rssi: 'Wi-Fi signal', color_temp: 'Colour temperature', battpercentage: 'Battery',
  illuminance_lux: 'Illuminance', current_heating_setpoint: 'Setpoint', local_temperature: 'Temperature',
  energy_today: 'Energy today', energy_yesterday: 'Energy yesterday', energy_month: 'Energy this month',
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

export const isWritable = (e: Expose): boolean => e.access_mode !== ExposeAccessModes.Read;

export function exposeKind(name: string): ExposeKind {
  if (KIND_BY_NAME[name]) return KIND_BY_NAME[name];
  if (name.startsWith('energy')) return 'energy';
  return 'misc';
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
export function isBinaryOn(e: Expose, value: unknown = e.data): boolean {
  return value === binaryOnValue(e);
}

export type NumericRange = { min: number; max: number };

/** Numeric bounds from expose attributes (hub uses min/max; Zigbee2MQTT value_min/value_max). */
export function exposeRange(e: Expose): NumericRange | null {
  const a = e.attributes ?? {};
  const max = Number(a['max'] ?? a['value_max']);
  if (!Number.isFinite(max)) return null;
  const min = Number(a['min'] ?? a['value_min'] ?? 0);
  return { min: Number.isFinite(min) ? min : 0, max };
}

/** Named preset values (e.g. colour temperature "warm" → 454), excluding sentinel values. */
export function exposePresets(e: Expose): { label: string; value: unknown }[] {
  if (!e.values) return [];
  const range = exposeRange(e);
  return Object.entries(e.values)
    .filter(([, v]) => !(range && typeof v === 'number' && (v < range.min || v > range.max)))
    .map(([k, v]) => (e.type === ExposeTypes.Enum ? { label: humanise(String(v)), value: v } : { label: humanise(k), value: v }));
}

export function capabilityOf(e: Expose): ExposeCapability {
  if (e.category === ExposeCategories.Config) return 'config';
  const writable = isWritable(e);
  if (e.type === ExposeTypes.Binary) return writable ? 'switch' : 'binary-sensor';
  if (e.type === ExposeTypes.Numeric && writable && exposeRange(e)) return 'slider';
  if (writable && e.values && Object.keys(e.values).length > 0 && (e.type === ExposeTypes.Enum || e.type === ExposeTypes.Numeric)) return 'preset';
  if (e.name.startsWith('action')) return 'event';
  if (e.category === ExposeCategories.Diagnostic) return 'diagnostic';
  return 'reading';
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

const numberText = (v: number): string => (Number.isInteger(v) ? String(v) : String(parseFloat(v.toFixed(Math.abs(v) < 10 ? 2 : 1))));

/**
 * Display text for an expose value (defaults to its current data). Uses expose metadata when
 * present: brightness becomes a percentage of its maximum; binary values become words.
 * Missing data is "—", never a fake zero (NH-06).
 */
export function formatExposeValue(e: Expose, value: unknown = e.data): string {
  if (value === null || value === undefined || value === '') return '—';
  if (e.type === ExposeTypes.Binary) {
    const [on, off] = BINARY_WORDS[e.name] ?? ['On', 'Off'];
    return isBinaryOn(e, value) ? on : off;
  }
  if (typeof value === 'number') {
    if (e.name === 'brightness') {
      const max = exposeRange(e)?.max || 254;
      return `${Math.round((value / max) * 100)} %`;
    }
    if (e.name === 'color_temp' && value > 0) return `${Math.round(1e6 / value)} K`;
    const unit = e.unit && e.unit !== 'lqi' ? e.unit : '';
    return unit ? `${numberText(value)} ${unit}` : numberText(value);
  }
  if (typeof value === 'string') {
    if (value === 'ON') return 'On';
    if (value === 'OFF') return 'Off';
    return humanise(value);
  }
  return getFormattedSensorValueByName(e.name, value, e.unit);
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
