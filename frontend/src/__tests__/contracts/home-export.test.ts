import { parseHomeExport, serializeHomeExport } from '@/contracts/home-export';
import type { HomeLayout } from '@/types/home.type';

const current = (): HomeLayout => ({
  version: 1,
  revision: 7,
  etag: 'current',
  blocks: [
    {
      id: 'area',
      kind: 'group',
      name: 'Kitchen',
      cards: [{ id: 'private', kind: 'waste_summary', instance_id: 'bins', resource_id: 'calendar', panel: false }],
    },
  ],
});

test('Home import uses the captured accepted revision, not metadata supplied by the file', () => {
  const exported = JSON.parse(serializeHomeExport({ ...current(), revision: 1, etag: 'old' }));
  const proposal = parseHomeExport(exported, current());
  expect(proposal).toMatchObject({ etag: 'current', revision: 7 });
  expect(proposal.blocks[0].cards[0]).toMatchObject({ id: 'private', panel: false });
  expect(Object.keys(exported.layout.blocks[0].cards[0]).sort()).toEqual([
    'id',
    'instance_id',
    'kind',
    'panel',
    'resource_id',
  ]);
});

test('legacy import preserves integration placements displaced by the device-only export', () => {
  const accepted = current();
  const proposal = parseHomeExport({ dashboardGroups: { Garden: { name: 'Garden', deviceGroup: {} } } }, accepted);
  expect(proposal.blocks[0].name).toBe('Garden');
  expect(proposal.blocks[1]).toMatchObject({ kind: 'card', cards: [{ id: 'private', panel: false }] });
  expect(accepted.blocks[0].name).toBe('Kitchen');
  expect(() => parseHomeExport({ format: 'unsupported', layout: accepted }, accepted)).toThrow();
  expect(() => parseHomeExport({ dashboardGroups: { Garden: null } }, accepted)).toThrow();
});
