/*
 * Protocol registry and device health helpers (spec 007 FR-07, AC-07…AC-09).
 * `connection_type` is set by the backend ingress path: "mqtt" (Zigbee2MQTT) or
 * "http" (POST /api/collect, e.g. Wi-Fi devices). Unknown values get a generic entry.
 */
import { Device } from '@/types/device';
import { DeviceAvailabilityTypes } from '@/types/device.type';

export type Diagnostic = 'lqi' | 'rssi' | 'battery';

export type ProtocolDescriptor = {
  id: string;
  label: string;
  /** Bridge operations offered for this protocol. */
  supports: { rename: boolean; interview: boolean; remove: boolean };
  generic: boolean;
};

const PROTOCOLS: Record<string, ProtocolDescriptor> = {
  mqtt: { id: 'mqtt', label: 'Zigbee', supports: { rename: true, interview: true, remove: true }, generic: false },
  http: { id: 'http', label: 'Wi-Fi · HTTP', supports: { rename: false, interview: false, remove: false }, generic: false },
};

export function resolveProtocol(connectionType: string | undefined | null): ProtocolDescriptor {
  const key = (connectionType ?? '').trim();
  return PROTOCOLS[key] ?? { id: key || 'unknown', label: key || 'Unknown', supports: { rename: false, interview: false, remove: false }, generic: true };
}

export const LOW_LINK_QUALITY = 50;
export const LOW_BATTERY = 20;

export const isOnline = (d: Device): boolean => d.availability === DeviceAvailabilityTypes.Online;
/** power_source is free text from the bridge, e.g. "battery", "mains (single phase)". */
export const isBatteryPowered = (d: Device): boolean => (d.power_source ?? '').toLowerCase().includes('battery');

const numeric = (v: unknown): number | null => (typeof v === 'number' && Number.isFinite(v) ? v : null);

export function batteryLevel(d: Device): number | null {
  return numeric(d.exposes?.['battery']?.data) ?? numeric(d.exposes?.['battpercentage']?.data);
}
export function linkQuality(d: Device): number | null {
  return numeric(d.exposes?.['linkquality']?.data);
}
export function wifiSignal(d: Device): number | null {
  return numeric(d.exposes?.['rssi']?.data);
}
export const hasLowBattery = (d: Device): boolean => {
  const b = batteryLevel(d);
  return b !== null && b < LOW_BATTERY;
};
export const hasWeakLink = (d: Device): boolean => {
  const l = linkQuality(d);
  return l !== null && l < LOW_LINK_QUALITY;
};

export type DeviceStatus = 'online' | 'offline' | 'disabled' | 'unknown';

/** Disabled (local setting) wins over availability so users see why values are hidden. */
export function deviceStatus(d: Device, disabled: boolean): DeviceStatus {
  if (disabled) return 'disabled';
  if (d.availability === DeviceAvailabilityTypes.Offline) return 'offline';
  if (d.availability === DeviceAvailabilityTypes.Online) return 'online';
  return 'unknown';
}

export function compareByName(a: Device, b: Device): number {
  return a.friendly_name.localeCompare(b.friendly_name, undefined, { sensitivity: 'base' });
}
