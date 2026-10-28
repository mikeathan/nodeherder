import type { HomeBlock, HomeCard, HomeCardSource, HomeLayout } from '@/types/home.type';
import { moveItem } from './dashboard';
import { generateConversationId } from '@/utils/unique';

/** Frontend bounds mirror the versioned server contract and are shared by editor/file ingress. */
export const HOME_LAYOUT_LIMITS = {
  blocks: 50,
  cards: 500,
  referenceBytes: 256,
  groupNameRunes: 60,
  fileBytes: 1_048_576,
} as const;

/** Source identity excludes placement ID, visibility, current value and translated labels. */
export function homeSourceKey(card: HomeCardSource): string {
  if (card.kind === 'device') return JSON.stringify([card.kind, card.device_id, card.expose]);
  return JSON.stringify([
    card.kind,
    card.instance_id,
    card.resource_id,
    card.kind === 'waste_service' ? card.service_id : '',
  ]);
}

/** Owned layout copy for a proposal; accepted arrays/cards must never be mutated by drag/import. */
export function cloneHome(layout: HomeLayout): HomeLayout {
  return {
    ...layout,
    blocks: layout.blocks.map((block) => ({ ...block, cards: block.cards.map((card) => ({ ...card })) })),
  };
}

/** Client-created presentation identity; contains no source details and survives data/label changes. */
export function newHomeId(): string {
  return generateConversationId();
}

/** Remove a placement, never its source. Empty named areas remain available for adding cards. */
export function withoutHomeCard(layout: HomeLayout, id: string): HomeLayout {
  const next = cloneHome(layout);
  for (const block of next.blocks) block.cards = block.cards.filter((card) => card.id !== id);
  next.blocks = next.blocks.filter((block) => block.kind === 'group' || block.cards.length > 0);
  return next;
}

/** Find a placement and its owning block by stable ID; absent cards return null. The returned
 * objects belong to the supplied snapshot and must be cloned before any mutation. */
export function findHomePlacement(
  blocks: readonly HomeBlock[],
  id: string
): { block: HomeBlock; card: HomeCard } | null {
  for (const block of blocks) {
    const card = block.cards.find((card) => card.id === id);
    if (card) return { block, card };
  }
  return null;
}

/** Move one reference atomically between a group and Home; standalone visibility-only edits do not reorder. */
export function moveHomeCard(
  layout: HomeLayout,
  id: string,
  destination: string | null,
  panel: boolean,
  index?: number
): HomeLayout {
  const next = cloneHome(layout);
  const placement = findHomePlacement(next.blocks, id);
  if (!placement) return next;
  const { block: owner, card } = placement;
  const target =
    destination === null ? null : next.blocks.find((block) => block.kind === 'group' && block.id === destination);
  if (destination !== null && !target) return next;
  card.panel = panel;
  if (owner === target || (owner.kind === 'card' && destination === null)) return next;
  owner.cards = owner.cards.filter((item) => item.id !== id);
  next.blocks = next.blocks.filter((block) => block.kind === 'group' || block.cards.length);
  if (target) {
    const requestedIndex = index ?? target.cards.length;
    const insertionIndex = Math.max(0, Math.min(requestedIndex, target.cards.length));
    target.cards.splice(insertionIndex, 0, card);
  } else {
    next.blocks.push({ id: newHomeId(), kind: 'card', name: '', cards: [card] });
  }
  return next;
}

/** Reorder blocks in reading order without dense-packing or changing per-card membership. */
export function moveHomeBlock(layout: HomeLayout, from: number, to: number): HomeLayout {
  const next = cloneHome(layout);
  next.blocks = moveItem(next.blocks, from, to);
  return next;
}

/** Validate a proposed area label against accepted groups; renaming excludes its own stable ID. */
export function homeAreaNameError(layout: HomeLayout, name: string, existingId?: string): string | null {
  const trimmed = name.trim();
  if (!trimmed) return 'Enter a name for the area.';
  if ([...trimmed].length > HOME_LAYOUT_LIMITS.groupNameRunes) return 'Use 60 characters or fewer.';
  if (/\p{Cc}/u.test(trimmed)) return 'Area names cannot contain control characters.';
  const duplicate = layout.blocks.some(
    (block) => block.kind === 'group' && block.id !== existingId && block.name.toLowerCase() === trimmed.toLowerCase()
  );
  return duplicate ? 'An area already has this name.' : null;
}

/** Create or rename an area on an owned proposal. Existing placement IDs remain unchanged. */
export function nameHomeArea(layout: HomeLayout, name: string, id?: string): HomeLayout {
  const next = cloneHome(layout);
  const group = next.blocks.find((block) => block.kind === 'group' && block.id === id);
  if (group) group.name = name.trim();
  else if (!id) next.blocks.push({ id: newHomeId(), kind: 'group', name: name.trim(), cards: [] });
  return next;
}

/** Remove one block and its placements from the proposal; no source is modified. */
export function withoutHomeBlock(layout: HomeLayout, id: string): HomeLayout {
  const next = cloneHome(layout);
  next.blocks = next.blocks.filter((block) => block.id !== id);
  return next;
}

/** Add owned placements to an area, deduplicating source choices there; null creates standalone blocks. */
export function addHomeCards(layout: HomeLayout, destination: string | null, cards: readonly HomeCard[]): HomeLayout {
  const next = cloneHome(layout);
  if (destination === null) {
    for (const card of cards) {
      next.blocks.push({ id: newHomeId(), kind: 'card', name: '', cards: [{ ...card }] });
    }
    return next;
  }
  const group = next.blocks.find((block) => block.kind === 'group' && block.id === destination);
  if (!group) return next;
  const existingSources = new Set(group.cards.map(homeSourceKey));
  for (const card of cards) {
    const source = homeSourceKey(card);
    if (!existingSources.has(source)) {
      group.cards.push({ ...card });
      existingSources.add(source);
    }
  }
  return next;
}

/** Reorder accepted blocks by ID. Invalid/transient drag membership cannot remove placements. */
export function reorderHomeBlocks(layout: HomeLayout, ids: readonly string[]): HomeLayout {
  const next = cloneHome(layout);
  if (ids.length !== next.blocks.length || new Set(ids).size !== ids.length) return next;
  const accepted = new Map(next.blocks.map((block) => [block.id, block]));
  const ordered: HomeBlock[] = [];
  for (const id of ids) {
    const block = accepted.get(id);
    if (!block) return next;
    ordered.push(block);
  }
  next.blocks = ordered;
  return next;
}

/** Reorder one area's existing card IDs; cross-area transfers use moveHomeCard as one transaction. */
export function reorderHomeCards(layout: HomeLayout, groupId: string, ids: readonly string[]): HomeLayout {
  const next = cloneHome(layout);
  const group = next.blocks.find((block) => block.kind === 'group' && block.id === groupId);
  if (!group || ids.length !== group.cards.length || new Set(ids).size !== ids.length) return next;
  const accepted = new Map(group.cards.map((card) => [card.id, card]));
  const ordered: HomeCard[] = [];
  for (const id of ids) {
    const card = accepted.get(id);
    if (!card) return next;
    ordered.push(card);
  }
  group.cards = ordered;
  return next;
}
