import type { HomeLayout } from '@/types/home.type';
import {
  addHomeCards,
  homeAreaNameError,
  nameHomeArea,
  reorderHomeBlocks,
  reorderHomeCards,
  withoutHomeBlock,
  homeSourceKey,
  moveHomeBlock,
  moveHomeCard,
  withoutHomeCard,
} from '@/domain/home-layout';

const layout = (): HomeLayout => ({
  version: 1,
  revision: 1,
  etag: '"layout"',
  blocks: [
    {
      id: 'kitchen',
      kind: 'group',
      name: 'Kitchen',
      cards: [{ id: 'lamp', kind: 'device', device_id: 'device', expose: 'state', panel: true }],
    },
    {
      id: 'standalone',
      kind: 'card',
      name: '',
      cards: [
        {
          id: 'black',
          kind: 'waste_service',
          instance_id: 'bins',
          resource_id: 'calendar',
          service_id: 'black',
          panel: true,
        },
      ],
    },
    { id: 'garden', kind: 'group', name: 'Garden', cards: [] },
  ],
});

test('moves individual cards atomically without mutating accepted source/order', () => {
  const current = layout();
  const before = structuredClone(current);
  const moved = moveHomeCard(current, 'black', 'garden', false);
  expect(moved.blocks.map((block) => block.id)).toEqual(['kitchen', 'garden']);
  expect(moved.blocks[1].cards[0]).toMatchObject({ id: 'black', panel: false, service_id: 'black' });
  expect(current).toEqual(before);
  expect(homeSourceKey(moved.blocks[1].cards[0])).toBe(homeSourceKey(current.blocks[1].cards[0]));
  const direct = moveHomeCard(moved, 'black', null, true);
  expect(direct.blocks.at(-1)).toMatchObject({ kind: 'card', name: '', cards: [{ id: 'black', panel: true }] });
});

test('visibility-only standalone changes never recreate or reorder the card block', () => {
  const current = layout();
  const next = moveHomeCard(current, 'black', null, false);
  expect(next.blocks.map((block) => block.id)).toEqual(current.blocks.map((block) => block.id));
  expect(next.blocks[1].cards[0].panel).toBe(false);
  expect(withoutHomeCard(current, 'black').blocks.map((block) => block.id)).toEqual(['kitchen', 'garden']);
  expect(moveHomeBlock(current, 0, 2).blocks.map((block) => block.id)).toEqual(['standalone', 'garden', 'kitchen']);
});

test('area/card commands own proposals and reject transient reorder membership', () => {
  const current = layout();
  const before = structuredClone(current);
  const card = current.blocks[0].cards[0];
  expect(addHomeCards(current, 'garden', [card, card]).blocks[2].cards).toHaveLength(1);
  expect(nameHomeArea(current, 'Lounge', 'kitchen').blocks[0]).toMatchObject({ id: 'kitchen', name: 'Lounge' });
  expect(reorderHomeBlocks(current, ['garden', 'standalone', 'kitchen']).blocks[0].id).toBe('garden');
  expect(reorderHomeBlocks(current, ['kitchen'])).toEqual(current);
  expect(reorderHomeBlocks(current, ['garden', 'garden', 'kitchen'])).toEqual(current);
  expect(reorderHomeCards(current, 'kitchen', ['unknown'])).toEqual(current);
  expect(withoutHomeBlock(current, 'kitchen').blocks.map((block) => block.id)).toEqual(['standalone', 'garden']);
  expect(homeAreaNameError(current, 'Garden')).toBe('An area already has this name.');
  expect(homeAreaNameError(current, 'Kitchen', 'kitchen')).toBeNull();
  expect(current).toEqual(before);
});
