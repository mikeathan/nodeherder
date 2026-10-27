import type { HomeBlock, HomeCard, HomeLayout } from '@/types/home.type';
import type { DashboardGroup, DashboardGroups } from '@/types/settings.type';
import { groupEntities, sortGroups } from './dashboard';
import { homeSourceKey, newHomeId } from './home-layout';

function legacyAreaBlock(group: DashboardGroup): HomeBlock {
  const cards: HomeCard[] = [];
  for (const [index, entity] of groupEntities(group).entries()) {
    cards.push({
      id: `fallback:${group.name}:${index}`,
      kind: 'device',
      panel: true,
      device_id: entity.deviceId,
      expose: entity.expose,
    });
  }
  return { id: `fallback:${group.name}`, kind: 'group', name: group.name, cards };
}

/** Preserve legacy device areas for read-only fallback when the protected endpoint is unavailable.
 * Temporary display IDs cannot be saved; canonical identity is supplied by the server migration. */
export function legacyHome(groups: DashboardGroups): HomeLayout {
  return { version: 1, revision: 0, etag: '', blocks: sortGroups(groups).map(legacyAreaBlock) };
}

/** Project a device-only compatibility area for existing panel climate/button semantics. */
export function homeDeviceGroup(block: HomeBlock): DashboardGroup {
  const group: DashboardGroup = { name: block.name, deviceGroup: Object.create(null) };
  for (const card of block.cards) {
    if (card.kind !== 'device') continue;
    const entry = group.deviceGroup[card.device_id] ?? { deviceId: card.device_id, exposes: [] };
    if (!entry.exposes.includes(card.expose)) entry.exposes.push(card.expose);
    group.deviceGroup[card.device_id] = entry;
  }
  return group;
}

/** Derive the legacy device-only view after an accepted read/save. Other screens see current
 * area counts without acquiring private integration references or another membership authority. */
export function homeDeviceGroups(layout: HomeLayout): DashboardGroups {
  const groups: DashboardGroups = Object.create(null);
  let order = 0;
  for (const block of layout.blocks) {
    if (block.kind === 'group') groups[block.name] = { ...homeDeviceGroup(block), order: order++ };
  }
  return groups;
}

/** Convert a device-only import without deleting private integration placements. Matching areas
 * retain placement identity/visibility; integration cards in displaced areas become standalone. */
export function importLegacyHome(layout: HomeLayout, groups: DashboardGroups): HomeLayout {
  const imported = legacyHome(groups);
  imported.etag = layout.etag;
  imported.revision = layout.revision;
  const previousAreas = new Map(
    layout.blocks.filter((block) => block.kind === 'group').map((block) => [block.name, block])
  );
  const importedNames = new Set(imported.blocks.map((block) => block.name));

  for (const area of imported.blocks) {
    const previous = previousAreas.get(area.name);
    area.id = previous?.id ?? newHomeId();
    const previousCards = new Map(previous?.cards.map((card) => [homeSourceKey(card), card]));
    for (const card of area.cards) {
      const accepted = previousCards.get(homeSourceKey(card));
      card.id = accepted?.id ?? newHomeId();
      card.panel = accepted?.panel ?? true;
    }
    if (previous) {
      for (const card of previous.cards) {
        if (card.kind !== 'device') area.cards.push({ ...card });
      }
    }
  }

  for (const block of layout.blocks) {
    if (block.kind === 'card') {
      if (block.cards[0]?.kind !== 'device')
        imported.blocks.push({ ...block, cards: block.cards.map((card) => ({ ...card })) });
      continue;
    }
    if (importedNames.has(block.name)) continue;
    for (const card of block.cards) {
      if (card.kind === 'device') continue;
      imported.blocks.push({ id: newHomeId(), kind: 'card', name: '', cards: [{ ...card }] });
    }
  }
  return imported;
}
