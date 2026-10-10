/*
 * Overview selectors (spec 007 FR-05, AC-09) and device-list filtering/sorting.
 * Pure functions of the device list; `isDisabled` is injected so this module never reads the store.
 */
import { Device } from '@/types/device';
import { batteryLevel, compareByName, deviceStatus, hasLowBattery, hasWeakLink, isBatteryPowered, isOnline, linkQuality } from './devices';

export type NetworkSummary = {
  total: number;
  online: number;
  offline: number;
  zigbeeRouters: number;
  zigbeeEndDevices: number;
  otherProtocols: number;
  weakLink: number;
  lowBattery: number;
};

export function summarizeNetwork(devices: Device[]): NetworkSummary {
  const zigbee = devices.filter((d) => d.connection_type === 'mqtt');
  const online = devices.filter(isOnline).length;
  return {
    total: devices.length,
    online,
    offline: devices.filter((d) => d.availability === 'offline').length,
    zigbeeRouters: zigbee.filter((d) => !isBatteryPowered(d)).length,
    zigbeeEndDevices: zigbee.filter(isBatteryPowered).length,
    otherProtocols: devices.length - zigbee.length,
    weakLink: devices.filter(hasWeakLink).length,
    lowBattery: devices.filter(hasLowBattery).length,
  };
}

export type AttentionReason = 'offline' | 'disabled' | 'low-battery' | 'weak-link';
export type AttentionItem = { device: Device; reason: AttentionReason };

const REASON_RANK: Record<AttentionReason, number> = { offline: 0, 'low-battery': 1, 'weak-link': 2, disabled: 3 };

/** One entry per device, most severe reason first, then by name. */
export function needsAttention(devices: Device[], isDisabled: (d: Device) => boolean): AttentionItem[] {
  const items: AttentionItem[] = [];
  for (const d of devices) {
    const status = deviceStatus(d, isDisabled(d));
    const reason: AttentionReason | null =
      status === 'offline' ? 'offline' : status === 'disabled' ? 'disabled' : hasLowBattery(d) ? 'low-battery' : hasWeakLink(d) ? 'weak-link' : null;
    if (reason) items.push({ device: d, reason });
  }
  return items.sort((a, b) => REASON_RANK[a.reason] - REASON_RANK[b.reason] || compareByName(a.device, b.device));
}

export type DeviceFilters = {
  query: string;
  protocol: 'all' | string;
  power: 'all' | 'battery' | 'mains';
  status: 'all' | 'online' | 'offline' | 'disabled';
};
export type DeviceSortKey = 'name' | 'signal' | 'battery' | 'lastSeen';
export const DEFAULT_FILTERS: DeviceFilters = { query: '', protocol: 'all', power: 'all', status: 'all' };

export function filterDevices(devices: Device[], f: DeviceFilters, isDisabled: (d: Device) => boolean): Device[] {
  const q = f.query.trim().toLowerCase();
  return devices.filter((d) => {
    if (q && !`${d.friendly_name} ${d.id} ${d.description ?? ''}`.toLowerCase().includes(q)) return false;
    if (f.protocol !== 'all' && d.connection_type !== f.protocol) return false;
    if (f.power !== 'all' && (f.power === 'battery') !== isBatteryPowered(d)) return false;
    if (f.status !== 'all') {
      const s = deviceStatus(d, isDisabled(d));
      if (f.status === 'online' ? s !== 'online' : s !== f.status) return false;
    }
    return true;
  });
}

/** Ascending = A–Z / weakest signal / lowest battery / most recent first. Missing values always sort last. */
export function sortDevices(devices: Device[], key: DeviceSortKey, ascending = true): Device[] {
  const value: Record<Exclude<DeviceSortKey, 'name'>, (d: Device) => number | null> = {
    signal: linkQuality,
    battery: batteryLevel,
    lastSeen: (d) => {
      const t = Date.parse(d.last_seen);
      return Number.isNaN(t) ? null : -t;
    },
  };
  const dir = ascending ? 1 : -1;
  return [...devices].sort((a, b) => {
    if (key === 'name') return dir * compareByName(a, b);
    const va = value[key](a), vb = value[key](b);
    if (va === null || vb === null) return va === vb ? compareByName(a, b) : va === null ? 1 : -1;
    return dir * (va - vb) || compareByName(a, b);
  });
}
