import type { HomeLayout } from '@/types/home.type';
import { moveHomeCard } from '@/domain/home-layout';
import { homePanelItems } from '@/domain/home-panel';

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

test('panel keeps mixed card order while retaining existing device control selection', () => {
  const current = moveHomeCard(layout(), 'black', 'kitchen', true, 0);
  const block = current.blocks[0];
  const button = { deviceId: 'device', expose: 'state', label: 'Lamp' };
  expect(
    homePanelItems(block, { name: 'Kitchen', climate: null, humidity: null, buttons: [button] }).map((item) => item.id)
  ).toEqual(['black', 'lamp']);
  block.cards[0].panel = false;
  expect(
    homePanelItems(block, { name: 'Kitchen', climate: null, humidity: null, buttons: [button] }).map((item) => item.id)
  ).toEqual(['lamp']);
  expect(homePanelItems(block, { name: 'Kitchen', climate: null, humidity: null, buttons: [] })).toEqual([]);
});
