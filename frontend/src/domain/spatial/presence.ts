import type { Device } from '@/types/device';
import { formatQuantity } from '@/utils/units';
import {
  derivedZoneSignals,
  type DerivedPresenceDisplay,
  type DerivedPresenceValues,
} from './derived';
import { supportsSpatialMapping } from './index';

/** User-facing copy for every presence surface; components render, they do not name. */
export const PRESENCE_COPY = {
  overview: {
    ariaLabel: 'Presence overview',
    sensorsAriaLabel: 'Presence sensors',
    noSensors: 'No spatial sensors are available yet.',
    remove: 'Remove presence overview',
    title: 'Presence',
    sensor: 'sensor',
    sensors: 'sensors',
  },
  card: {
    ariaLabel: (name: string) => `Presence for ${name}`,
    unknownDevice: 'Unknown device',
    remove: 'Remove presence card',
    missing: 'This presence device is unavailable. Remove the card if it is no longer needed.',
    notDerived: 'Presence has not been derived for this device yet.',
    notDerivedNote: 'The hub derives presence from target positions; it updates when this device reports one.',
    live: 'Live',
    livePreview: 'Live preview',
    previewLabel: 'Presence preview',
    waiting: 'Waiting for a position…',
    openDevice: 'Open device',
    zoneOccupancy: 'Zone occupancy',
  },
  tile: {
    ariaLabel: (name: string) => `${name} presence`,
    previewLabel: (name: string) => `${name} preview`,
    noPosition: 'No position yet',
    fromSensor: 'from sensor',
    actions: [
      { label: 'Open live', tab: 'presence' },
      { label: 'History', tab: 'history' },
    ],
  },
  state: {
    notDerived: 'Presence not derived yet',
    compactNotDerived: 'Not derived',
    occupied: 'Occupied',
    clear: 'Clear',
  },
  zone: {
    unknown: 'unknown',
    occupied: 'occupied',
    clear: 'clear',
  },
  labels: {
    target: 'target',
    targets: 'targets',
    lastZone: 'last zone',
    empty: 'empty',
  },
  server: {
    prefix: 'Server presence',
    notDerived: 'not derived yet',
    offline: 'derived device offline',
    noneYet: 'none yet',
  },
} as const;

const STATE_LABELS: Record<DerivedPresenceDisplay, { compact: string; card: string }> = {
  'not-derived': { compact: PRESENCE_COPY.state.compactNotDerived, card: PRESENCE_COPY.state.notDerived },
  occupied: { compact: PRESENCE_COPY.state.occupied, card: PRESENCE_COPY.state.occupied },
  clear: { compact: PRESENCE_COPY.state.clear, card: PRESENCE_COPY.state.clear },
};

/** Label for a derived presence state; cards use the explicit not-derived wording. */
export function presenceStateLabel(display: DerivedPresenceDisplay, variant: 'compact' | 'card' = 'compact'): string {
  return STATE_LABELS[display][variant];
}

/** "2 targets" or null when the count is unknown. */
export function presenceTargetLabel(count: number | null): string | null {
  if (count === null) return null;
  return `${count} ${count === 1 ? PRESENCE_COPY.labels.target : PRESENCE_COPY.labels.targets}`;
}

/** "last zone Desk" or null when no zone was reported. */
export function presenceZoneLabel(zone: string | null): string | null {
  return zone ? `${PRESENCE_COPY.labels.lastZone} ${zone}` : null;
}

/** "empty 12 s" or null when the duration is unknown. */
export function presenceEmptyLabel(seconds: number | null): string | null {
  return seconds === null ? null : `${PRESENCE_COPY.labels.empty} ${seconds} s`;
}

/** "1.5 m from sensor" or "No position yet". */
export function presenceDistanceLabel(distanceMm: number | null): string {
  if (distanceMm === null || !Number.isFinite(distanceMm)) return PRESENCE_COPY.tile.noPosition;
  return `${formatQuantity(distanceMm / 1000, 'distance', { digits: 1 })} ${PRESENCE_COPY.tile.fromSensor}`;
}

/** Word for a zone's occupancy, distinct from the global presence state. */
export function presenceZoneOccupancyLabel(occupied: boolean | null): string {
  if (occupied === null) return PRESENCE_COPY.zone.unknown;
  return occupied ? PRESENCE_COPY.zone.occupied : PRESENCE_COPY.zone.clear;
}

/** One derived zone chip for the full presence card. */
export type PresenceZoneChip = { id: string; name: string; occupied: boolean | null; label: string };

/** Resolves zone names and labels without duplicating the entity walk in a component. */
export function presenceZoneChips(device: Device | undefined): PresenceZoneChip[] {
  const signals = derivedZoneSignals(device);
  return [...signals.entries()]
    .map(([zoneId, signal]) => {
      const entity = Object.values(device?.exposes ?? {}).find(
        (candidate) => candidate.attributes?.zoneId === zoneId && candidate.name.endsWith('_occupied')
      );
      const name = typeof entity?.attributes?.zoneName === 'string' ? entity.attributes.zoneName : zoneId;
      return { id: zoneId, name, occupied: signal.occupied, label: presenceZoneOccupancyLabel(signal.occupied) };
    })
    .sort((left, right) => left.name.localeCompare(right.name));
}

/** Header and list model for the Home presence overview card. */
export type PresenceOverviewModel = {
  sensors: Device[];
  title: string;
  subtitle: string;
  showHeader: boolean;
  showNames: boolean;
  isSingle: boolean;
};

/**
 * A single sensor is the card itself, so no second header is drawn. Multiple sensors need
 * one section header and named rows. Editing keeps a header for the remove control.
 */
export function presenceOverviewModel(
  devices: readonly Device[],
  editing: boolean,
  statusOf: (device: Device) => string
): PresenceOverviewModel {
  const sensors = devices.filter(supportsSpatialMapping);
  const single = sensors.length === 1 ? sensors[0] : undefined;
  return {
    sensors,
    title: single?.friendly_name ?? PRESENCE_COPY.overview.title,
    subtitle: single
      ? statusOf(single)
      : `${sensors.length} ${sensors.length === 1 ? PRESENCE_COPY.overview.sensor : PRESENCE_COPY.overview.sensors}`,
    showHeader: !single || editing,
    showNames: !single || !editing,
    isSingle: !!single,
  };
}

/** One-line readout for the live spatial panel; missing data is never called clear. */
export function serverPresenceLabel(input: {
  hasDerivedDevice: boolean;
  online: boolean;
  values: DerivedPresenceValues;
}): string {
  const prefix = `${PRESENCE_COPY.server.prefix}:`;
  if (!input.hasDerivedDevice) return `${prefix} ${PRESENCE_COPY.server.notDerived}.`;
  if (!input.online) return `${prefix} ${PRESENCE_COPY.server.offline}.`;
  if (input.values.occupied === null) return `${prefix} ${PRESENCE_COPY.server.notDerived}.`;
  const count = input.values.targetCount ?? 0;
  const empty = input.values.emptySeconds ?? 0;
  const last = input.values.lastZone ?? PRESENCE_COPY.server.noneYet;
  const targets = count === 1 ? PRESENCE_COPY.labels.target : PRESENCE_COPY.labels.targets;
  return `${prefix} ${input.values.occupied ? PRESENCE_COPY.zone.occupied : PRESENCE_COPY.zone.clear} · ${count} ${targets} · ${PRESENCE_COPY.labels.lastZone} ${last} · ${PRESENCE_COPY.labels.empty} ${empty}s`;
}
