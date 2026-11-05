/*
 * Device page view models: tabs, category sections, connection copy and About
 * rows. Keeping this out of DevicePage.vue makes the component presentational.
 */
import type { Device, Expose } from '@/types/device';
import { ExposeCategories } from '@/types/device.type';
import { isESPHomeDevice, resolveProtocol } from './index';
import { isWritable } from '@/domain/exposes';
import { supportsSpatialMapping } from '@/domain/spatial';

/** Stable tab identifiers. */
export const DeviceTabIds = {
  Controls: 'controls',
  Presence: 'presence',
  History: 'history',
  Settings: 'settings',
  About: 'about',
} as const;

/** Tabs in display order. Presence is offered only to devices that can map coordinates. */
export const DEVICE_TABS = [
  { id: DeviceTabIds.Controls, label: 'Controls' },
  { id: DeviceTabIds.Presence, label: 'Presence' },
  { id: DeviceTabIds.History, label: 'History' },
  { id: DeviceTabIds.Settings, label: 'Settings' },
  { id: DeviceTabIds.About, label: 'About' },
] as const;

/** One device-page tab. */
export type DeviceTab = (typeof DEVICE_TABS)[number];

/**
 * Tabs available for one device. Presence is withheld unless the device exposes mappable
 * coordinates and an active flag, so unrelated devices gain no empty tab.
 */
export function deviceTabs(device: Device | undefined): readonly DeviceTab[] {
  return DEVICE_TABS.filter((tab) => tab.id !== DeviceTabIds.Presence || supportsSpatialMapping(device));
}

/** Stable tab identifier used in the URL. */
export type DeviceTabId = (typeof DEVICE_TABS)[number]['id'];

/** Provider-reported metadata shown in the About tab. */
export const DEVICE_PROVIDER_INFO_FIELDS = [
  { key: 'manufacturer', label: 'Manufacturer' },
  { key: 'firmware_version', label: 'ESPHome version' },
  { key: 'project_name', label: 'Project' },
  { key: 'project_version', label: 'Project version' },
] as const;

/** Placeholder for an absent About value. */
export const DEVICE_ABOUT_MISSING = '—';

/** Resolves an arbitrary query value to an available tab; unknown or withheld falls back. */
export function resolveDeviceTab(value: unknown, tabs: readonly DeviceTab[] = DEVICE_TABS): DeviceTabId {
  return tabs.some((tab) => tab.id === value) ? (value as DeviceTabId) : DeviceTabIds.Controls;
}

/** URL query value for a tab; the default tab omits the parameter. */
export function deviceTabQuery(tab: DeviceTabId): DeviceTabId | undefined {
  return tab === DeviceTabIds.Controls ? undefined : tab;
}

/** Tab label, with the ESPHome-specific hub settings wording. */
export function deviceTabLabel(tab: DeviceTabId, isESPHome: boolean): string {
  if (isESPHome && tab === DeviceTabIds.Settings) return 'Hub settings';
  return DEVICE_TABS.find((item) => item.id === tab)?.label ?? tab;
}

/** Wraps keyboard navigation across the tabs actually shown for this device. */
export function adjacentDeviceTab(tabs: readonly DeviceTab[], index: number, step: number): DeviceTab {
  return tabs[(index + step + tabs.length) % tabs.length];
}

/** Entities in one category for the given device. */
export function deviceExposesByCategory(device: Device | undefined, category: string): Expose[] {
  return Object.values(device?.exposes ?? {}).filter((expose) => expose.category === category);
}

/** Label for the last report row; provider devices report rather than arrive. */
export function lastSeenLabel(device: Device | undefined): string {
  return isESPHomeDevice(device) ? 'Last report' : 'Last seen';
}

/** One About row. `relative` renders the shared relative-time component. */
export type DeviceAboutRow = {
  key: string;
  label: string;
  value?: string;
  code?: boolean;
  relative?: boolean;
  note?: string;
};

function publicProperty(device: Device, key: string): string {
  const value = device.properties?.[key];
  return typeof value === 'string' && value.trim() ? value : DEVICE_ABOUT_MISSING;
}

/** Builds the About rows without exposing credentials or network addresses. */
export function deviceAboutRows(device: Device): DeviceAboutRow[] {
  const protocol = resolveProtocol(device.connection_type);
  const rows: DeviceAboutRow[] = [
    { key: 'name', label: 'Name', value: device.friendly_name },
    { key: 'address', label: 'Address', value: device.id, code: true },
    { key: 'model', label: 'Model', value: device.properties?.model || device.description || DEVICE_ABOUT_MISSING },
  ];
  if (isESPHomeDevice(device)) {
    for (const field of DEVICE_PROVIDER_INFO_FIELDS) {
      rows.push({ key: field.key, label: field.label, value: publicProperty(device, field.key) });
    }
  }
  rows.push(
    {
      key: 'connection',
      label: 'Connection',
      value: protocol.label,
      note: protocol.generic ? ` (reported as “${device.connection_type || 'unknown'}”)` : undefined,
    },
    { key: 'power', label: 'Power', value: device.power_source || DEVICE_ABOUT_MISSING },
    { key: 'availability', label: 'Availability', value: device.availability },
    { key: 'last_seen', label: lastSeenLabel(device), relative: true }
  );
  return rows;
}

/** Measurement, configuration and diagnostic sections for a device. */
export function deviceCategorySections(device: Device | undefined) {
  const measurements = deviceExposesByCategory(device, ExposeCategories.Measurement);
  return {
    measurements,
    readings: measurements.filter((expose) => !isWritable(expose)),
    controls: measurements.filter(isWritable),
    configuration: deviceExposesByCategory(device, ExposeCategories.Config),
    diagnostics: deviceExposesByCategory(device, ExposeCategories.Diagnostic),
  };
}
