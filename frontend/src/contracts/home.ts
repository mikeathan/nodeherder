import { parseIntegrationReference } from '@/integrations/registry';
import type { HomeBlock, HomeCard, HomeLayout } from '@/types/home.type';
import type { DashboardGroup, DashboardGroups } from '@/types/settings.type';
import { HOME_LAYOUT_LIMITS } from '@/domain/home-layout';

const encoder = new TextEncoder();

function readRecord(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('Invalid Home object');
  return value as Record<string, unknown>;
}

function readBoundedText(value: unknown, maxBytes: number, allowEmpty = false): string {
  if (typeof value !== 'string' || (!allowEmpty && !value) || /\p{Cc}/u.test(value)) {
    throw new Error('Invalid Home text');
  }
  if (encoder.encode(value).length > maxBytes) throw new Error('Home text exceeds its limit');
  return value;
}

function readReference(value: unknown): string {
  return readBoundedText(value, HOME_LAYOUT_LIMITS.referenceBytes);
}

function readAreaName(value: unknown): string {
  const name = readBoundedText(value, HOME_LAYOUT_LIMITS.groupNameRunes * 4);
  if (name.trim() !== name || [...name].length > HOME_LAYOUT_LIMITS.groupNameRunes) {
    throw new Error('Invalid Home group name');
  }
  return name;
}

function rejectMixedReferences(raw: Record<string, unknown>, excluded: readonly string[]) {
  if (excluded.some((field) => raw[field] !== undefined && raw[field] !== '')) throw new Error('Mixed Home references');
}

function readHomeCard(value: unknown): HomeCard {
  const raw = readRecord(value);
  const id = readReference(raw.id);
  if (typeof raw.panel !== 'boolean') throw new Error('Invalid panel visibility');
  const panel = raw.panel;
  switch (raw.kind) {
    case 'device':
      rejectMixedReferences(raw, ['instance_id', 'resource_id', 'service_id']);
      return { id, panel, kind: 'device', device_id: readReference(raw.device_id), expose: readReference(raw.expose) };
    default: {
      rejectMixedReferences(raw, ['device_id', 'expose']);
      readReference(raw.instance_id);
      readReference(raw.resource_id);
      if (raw.service_id) readReference(raw.service_id);
      return { id, panel, ...parseIntegrationReference(raw) };
    }
  }
}

interface LayoutMembership {
  blockIds: Set<string>;
  cardIds: Set<string>;
  areaNames: Set<string>;
}

function readHomeBlock(value: unknown, membership: LayoutMembership): HomeBlock {
  const raw = readRecord(value);
  const id = readReference(raw.id);
  if (membership.blockIds.has(id) || !Array.isArray(raw.cards) || raw.cards.length > HOME_LAYOUT_LIMITS.cards) {
    throw new Error('Invalid Home block');
  }
  membership.blockIds.add(id);
  const cards: HomeCard[] = [];
  for (const value of raw.cards) {
    const card = readHomeCard(value);
    if (membership.cardIds.has(card.id) || membership.cardIds.size >= HOME_LAYOUT_LIMITS.cards) {
      throw new Error('Duplicate or excessive Home cards');
    }
    membership.cardIds.add(card.id);
    cards.push(card);
  }
  if (raw.kind === 'card' && cards.length === 1 && raw.name === '') return { id, kind: 'card', name: '', cards };
  if (raw.kind !== 'group') throw new Error('Invalid Home block kind');
  const name = readAreaName(raw.name);
  const key = name.toLowerCase();
  if (membership.areaNames.has(key)) throw new Error('Duplicate Home group name');
  membership.areaNames.add(key);
  return { id, kind: 'group', name, cards };
}

/** Decode an owned layout at ingress. Reject malformed/duplicate/oversized membership before it
 * reaches Vuex or an import proposal; false visibility and empty arrays retain their wire meaning. */
export function parseHomeLayout(value: unknown): HomeLayout {
  const raw = readRecord(value);
  if (raw.version !== 1 || !Number.isSafeInteger(raw.revision) || Number(raw.revision) < 0) {
    throw new Error('Unsupported Home layout');
  }
  if (!Array.isArray(raw.blocks) || raw.blocks.length > HOME_LAYOUT_LIMITS.blocks)
    throw new Error('Invalid Home blocks');
  const membership: LayoutMembership = { blockIds: new Set(), cardIds: new Set(), areaNames: new Set() };
  const blocks = raw.blocks.map((value) => readHomeBlock(value, membership));
  const etag = readBoundedText(raw.etag, HOME_LAYOUT_LIMITS.referenceBytes, true);
  return { version: 1, revision: Number(raw.revision), etag, blocks };
}

function readLegacyDevice(id: string, value: unknown): DashboardGroup['deviceGroup'][string] {
  const raw = readRecord(value);
  if (readReference(raw.deviceId) !== id || !Array.isArray(raw.exposes))
    throw new Error('Invalid imported device values');
  if (raw.exposes.length > HOME_LAYOUT_LIMITS.cards) throw new Error('Too many imported values');
  const exposes = raw.exposes.map(readReference);
  if (new Set(exposes).size !== exposes.length) throw new Error('Duplicate imported values');
  return { deviceId: id, exposes };
}

function readLegacyArea(key: string, value: unknown): DashboardGroup {
  const raw = readRecord(value);
  const name = readAreaName(raw.name);
  if (name !== key) throw new Error('Invalid imported area name');
  const deviceGroup: DashboardGroup['deviceGroup'] = Object.create(null);
  for (const [id, device] of Object.entries(readRecord(raw.deviceGroup)))
    deviceGroup[id] = readLegacyDevice(id, device);
  const order = raw.order;
  if (order === undefined || order === null) return { name, deviceGroup };
  if (!Number.isSafeInteger(order) || Number(order) < 0) throw new Error('Invalid imported area order');
  return { name, deviceGroup, order: Number(order) };
}

/** Validate device-only legacy exports before compatibility conversion. Imported arrays/maps are
 * owned, bounded and prototype-safe; order is an optional nonnegative integer, not a UI index. */
export function parseLegacyHomeGroups(value: unknown): DashboardGroups {
  const entries = Object.entries(readRecord(value));
  if (entries.length > HOME_LAYOUT_LIMITS.blocks) throw new Error('Too many imported areas');
  const groups: DashboardGroups = Object.create(null);
  const names = new Set<string>();
  let cardCount = 0;
  for (const [key, value] of entries) {
    const group = readLegacyArea(key, value);
    const nameKey = group.name.toLowerCase();
    if (names.has(nameKey)) throw new Error('Duplicate imported area name');
    names.add(nameKey);
    for (const device of Object.values(group.deviceGroup)) cardCount += device.exposes.length;
    if (cardCount > HOME_LAYOUT_LIMITS.cards) throw new Error('Too many imported values');
    groups[group.name] = group;
  }
  return groups;
}
