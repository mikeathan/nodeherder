import type { HomeBlock, HomeIntegrationCard, HomeLayout } from '@/types/home.type';
import type { PanelButton, PanelRoom } from './panel';

export type HomePanelItem =
  | { kind: 'device'; id: string; button: PanelButton }
  | { kind: 'integration'; id: string; card: HomeIntegrationCard };

/** Preserve mixed placement order around the existing climate/device-button projection.
 * The panel domain still selects/deduplicates device controls; this adapter only orders them. */
export function homePanelItems(block: HomeBlock, room: PanelRoom): HomePanelItem[] {
  const buttons = new Map(room.buttons.map((button) => [JSON.stringify([button.deviceId, button.expose]), button]));
  const items: HomePanelItem[] = [];
  for (const card of block.cards) {
    if (!card.panel) continue;
    if (card.kind !== 'device') {
      items.push({ kind: 'integration', id: card.id, card });
      continue;
    }
    const key = JSON.stringify([card.device_id, card.expose]);
    const button = buttons.get(key);
    if (button) {
      items.push({ kind: 'device', id: card.id, button });
      buttons.delete(key);
    }
  }
  return items;
}

/** Select visible panel pages in accepted order, leaving the Home layout untouched. */
export function homePanelBlocks(layout: HomeLayout): HomeBlock[] {
  const pages: HomeBlock[] = [];
  for (const block of layout.blocks) {
    const cards = block.cards.filter((card) => card.panel);
    if (cards.length > 0) pages.push({ ...block, cards });
  }
  return pages;
}
