import type { HubStateSnapshot } from '@/types/hub.type';

function record(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

/** Validate the registry/config containers before they can replace shared state. */
export function isHubStateSnapshot(value: unknown): value is HubStateSnapshot {
  if (!record(value) || !Array.isArray(value.devices) || !record(value.config)) return false;
  if (!record(value.config.hub) || !record(value.config.bridge)) return false;
  const hub = value.config.hub;
  if (!record(hub.devices) || !record(hub.devices.defaults) || !record(hub.devices.overrides)) return false;
  if (![hub.history, hub.logger, hub.assistant, hub.mcp, hub.dashboardGroups].every(record)) return false;
  return value.devices.every((device: unknown) => record(device) && typeof device.id === 'string' &&
    device.id.length > 0 && record(device.exposes) && Object.values(device.exposes).every(record));
}
