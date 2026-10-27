import type { HomeLayout } from '@/types/home.type';
import { cloneHome, moveHomeCard } from '@/domain/home-layout';
import { homeDeviceGroup, importLegacyHome } from '@/domain/home-compatibility';

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

test('legacy import preserves displaced integration placement and device-only projection excludes it', () => {
  const current = moveHomeCard(layout(), 'black', 'kitchen', false);
  const result = importLegacyHome(current, { Lounge: { name: 'Lounge', deviceGroup: {} } });
  expect(result.blocks[0].name).toBe('Lounge');
  expect(result.blocks[1]).toMatchObject({ kind: 'card', cards: [{ id: 'black', panel: false }] });
  const copied = cloneHome(current);
  copied.blocks[0].cards[0].panel = false;
  expect(current.blocks[0].cards[0].panel).toBe(true);
  expect(homeDeviceGroup(current.blocks[0]).deviceGroup).toEqual({
    device: { deviceId: 'device', exposes: ['state'] },
  });
});
