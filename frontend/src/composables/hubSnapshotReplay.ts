import type { MutationPayload } from 'vuex';
import { store } from '@/store';
import { appConfigMutations } from '@/store/modules/hub-state';
import type { Device, DeviceUpdate } from '@/types/device';
import type { HubStateSnapshot } from '@/types/hub.type';

const maxConfigEdits = 256;
const maxEarlyDevices = 256;
const maxEarlyFields = 256;
const hasOwn = (value: object, key: string) => Object.prototype.hasOwnProperty.call(value, key);

function mergeReport(device: Device, update: DeviceUpdate): Device {
  const exposes = { ...device.exposes };
  for (const [name, value] of Object.entries(update.data)) {
    if (hasOwn(exposes, name)) exposes[name] = { ...exposes[name], data: value };
  }
  return { ...device, exposes, last_seen: update.last_seen || device.last_seen, availability: update.availability || device.availability };
}

function reportFields(update: DeviceUpdate, previous: DeviceUpdate | undefined, device: Device | undefined): Record<string, unknown> {
  const fields: Record<string, unknown> = Object.assign(Object.create(null), previous?.data);
  const incoming = update.data && typeof update.data === 'object' && !Array.isArray(update.data) ? update.data : {};
  let count = Object.keys(fields).length;
  for (const [name, value] of Object.entries(incoming)) {
    if (device && !hasOwn(device.exposes, name)) continue;
    const isNew = !hasOwn(fields, name);
    if (!device && (name.length > 256 || (isNew && count >= maxEarlyFields))) throw new Error('Early report field limit exceeded');
    fields[name] = value;
    if (isNew) count++;
  }
  return fields;
}

/** Captures bounded Vuex changes during one HTTP snapshot. The caller owns stop().
 * Reconciliation copies the server snapshot; unknown reports never create devices.
 * Overflow calls onOverflow synchronously so the request can abort and unsubscribe. */
export function createHubSnapshotReplay(onOverflow: () => void) {
  const updates = new Map<string, DeviceUpdate>();
  const added = new Map<string, Device>();
  const earlyDevices = new Set<string>();
  const configEdits: MutationPayload[] = [];
  let listChanged = false;

  function reserveEarlyDevice(id: string) {
    if (earlyDevices.has(id)) return;
    if (earlyDevices.size >= maxEarlyDevices) throw new Error('Early device limit exceeded');
    earlyDevices.add(id);
  }

  function recordReport(update: DeviceUpdate) {
    if (!update || typeof update.id !== 'string' || !update.id || update.id.length > 256) return;
    const device = store.getters['hub/findDevice'](update.id) as Device | undefined;
    if (!device) reserveEarlyDevice(update.id);
    const previous = updates.get(update.id);
    const data = reportFields(update, previous, device);
    updates.set(update.id, {
      ...previous, ...update, data,
      last_seen: update.last_seen || previous?.last_seen || '',
      availability: update.availability || previous?.availability || null,
    });
  }

  function recordConfigEdit(mutation: MutationPayload) {
    if (configEdits.length >= maxConfigEdits) throw new Error('Configuration replay limit exceeded');
    // Store mutations may retain mutable payloads; replay needs the value at receipt.
    configEdits.push({ type: mutation.type.slice(4), payload: mutation.payload === undefined ? undefined : JSON.parse(JSON.stringify(mutation.payload)) });
  }

  function recordMutation(mutation: MutationPayload) {
    if (mutation.type === 'hub/updateDevice') return recordReport(mutation.payload as DeviceUpdate);
    if (mutation.type === 'hub/addDevice') {
      const device = mutation.payload as Device;
      added.set(device.id, device);
      return;
    }
    if (mutation.type === 'hub/setDevices') { listChanged = true; return; }
    if (mutation.type.startsWith('hub/') && hasOwn(appConfigMutations, mutation.type.slice(4))) recordConfigEdit(mutation);
  }

  const unsubscribe = store.subscribe((mutation) => {
    try { recordMutation(mutation); }
    catch { onOverflow(); }
  });

  function reconcileDevices(snapshot: HubStateSnapshot): Device[] {
    if (listChanged) return store.getters['hub/listAllDevices']() as Device[];
    const devices = new Map(snapshot.devices.map((device) => [device.id, device]));
    for (const [id, device] of added) devices.set(id, device);
    for (const [id, update] of updates) {
      const device = devices.get(id);
      if (device) devices.set(id, mergeReport(device, update));
    }
    return [...devices.values()];
  }

  function reconcileConfig(config: HubStateSnapshot['config']): HubStateSnapshot['config'] {
    if (!configEdits.length) return config;
    const next = { appConfig: JSON.parse(JSON.stringify(config)) as HubStateSnapshot['config'] };
    for (const edit of configEdits) appConfigMutations[edit.type](next, edit.payload);
    return next.appConfig;
  }

  /** Replays captured observations and configuration edits over a server snapshot without mutating it. */
  function reconcile(snapshot: HubStateSnapshot): HubStateSnapshot {
    return { ...snapshot, devices: reconcileDevices(snapshot), config: reconcileConfig(snapshot.config) };
  }

  /** Stops observing this attempt; safe on abort and again after settlement. */
  function stop() { unsubscribe(); }

  return { reconcile, stop };
}
