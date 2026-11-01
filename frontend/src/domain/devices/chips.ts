/*
 * Presentation view models for the compact device status chips.
 * Formatting and threshold decisions live here so components only render.
 */
import type { Device } from '@/types/device';
import { batteryLevel, hasLowBattery, hasWeakLink, isBatteryPowered, linkQuality, wifiSignal } from './index';
import { formatQuantity, formatWithUnit } from '@/utils/units';

/** Chip tones supported by the shared chip component. */
export type DeviceChipTone = 'default' | 'warn' | 'muted';

/** Battery label when a device reports power but no measurable charge. */
export const BATTERY_WITHOUT_LEVEL_LABEL = 'Battery';
/** Label for a device that reports a mains power source. */
export const MAINS_POWER_LABEL = 'Mains';
/** Title describing the Zigbee link-quality unit. */
export const LINK_QUALITY_TITLE = 'Link quality (0–255)';
/** Title describing the Wi-Fi signal unit. */
export const WIFI_SIGNAL_TITLE = 'Wi-Fi signal';

/** Link-quality or Wi-Fi signal chip, or null when the device reports neither. */
export function linkChip(device: Device): { icon: 'signal' | 'wifi'; label: string; tone: DeviceChipTone; title: string } | null {
  const quality = linkQuality(device);
  if (quality !== null) {
    return {
      icon: 'signal',
      label: formatWithUnit(quality, 'LQI'),
      tone: hasWeakLink(device) ? 'warn' : 'default',
      title: LINK_QUALITY_TITLE,
    };
  }
  const signal = wifiSignal(device);
  if (signal !== null) {
    return { icon: 'wifi', label: formatWithUnit(signal, 'dBm'), tone: 'default', title: WIFI_SIGNAL_TITLE };
  }
  return null;
}

/** Power chip for battery or mains devices, or null when no source is reported. */
export function powerChip(device: Device): { icon?: 'battery' | 'batteryLow'; label: string; tone: DeviceChipTone } | null {
  if (isBatteryPowered(device)) {
    const level = batteryLevel(device);
    const low = hasLowBattery(device);
    return {
      icon: low ? 'batteryLow' : 'battery',
      label: level === null ? BATTERY_WITHOUT_LEVEL_LABEL : formatQuantity(level, 'percent', { spaced: true }),
      tone: low ? 'warn' : 'default',
    };
  }
  return device.power_source ? { label: MAINS_POWER_LABEL, tone: 'muted' } : null;
}
