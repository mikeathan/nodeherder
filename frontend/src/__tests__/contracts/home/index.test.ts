import { parseHomeLayout, parseLegacyHomeGroups } from '@/contracts/home';

const valid = () => ({
  version: 1,
  revision: 0,
  etag: '"test"',
  blocks: [
    {
      id: 'kitchen',
      kind: 'group',
      name: 'Kitchen',
      cards: [{ id: 'one', kind: 'device', panel: true, device_id: 'lamp', expose: 'state' }],
    },
  ],
});

test('layout parsing owns references and preserves false panel choices and empty arrays', () => {
  const raw = valid();
  raw.blocks[0].cards[0].panel = false;
  const layout = parseHomeLayout(raw);
  layout.blocks[0].cards[0].panel = true;
  expect(raw.blocks[0].cards[0].panel).toBe(false);
  expect(parseHomeLayout({ version: 1, revision: 0, etag: '', blocks: [] }).blocks).toEqual([]);
});

test('text ingress uses server byte bounds and full Unicode control checks', () => {
  const raw = valid();
  raw.blocks[0].name = '🏠'.repeat(60);
  expect(parseHomeLayout(raw).blocks[0].name).toBe(raw.blocks[0].name);
  raw.blocks[0].id = '🏠'.repeat(65);
  expect(() => parseHomeLayout(raw)).toThrow();
  raw.blocks[0].id = 'group';
  raw.blocks[0].name = 'Kitchen\u0085';
  expect(() => parseHomeLayout(raw)).toThrow();
});

test.each([
  null,
  {},
  { ...valid(), blocks: null },
  { ...valid(), version: 2 },
  { ...valid(), revision: 1.5 },
  { ...valid(), blocks: [...valid().blocks, ...valid().blocks] },
  { ...valid(), blocks: [{ ...valid().blocks[0], cards: [{ ...valid().blocks[0].cards[0], panel: null }] }] },
  {
    ...valid(),
    blocks: [{ ...valid().blocks[0], cards: [{ ...valid().blocks[0].cards[0], instance_id: 'private' }] }],
  },
  { ...valid(), blocks: [{ ...valid().blocks[0], cards: [{ ...valid().blocks[0].cards[0], kind: 'script' }] }] },
])('rejects malformed/ambiguous wire state %p', (value) => {
  expect(() => parseHomeLayout(value)).toThrow();
});

test('legacy conversion ingress rejects null, duplicate values and unsafe order before mutation', () => {
  const good = {
    Kitchen: { name: 'Kitchen', deviceGroup: { lamp: { deviceId: 'lamp', exposes: ['state'] } }, order: 0 },
  };
  expect(parseLegacyHomeGroups(good)).toEqual(good);
  for (const value of [
    null,
    { Kitchen: null },
    { Kitchen: { ...good.Kitchen, order: 0.5 } },
    { Kitchen: { ...good.Kitchen, deviceGroup: { lamp: { deviceId: 'other', exposes: ['state'] } } } },
  ])
    expect(() => parseLegacyHomeGroups(value)).toThrow();
});

test.each([null, false, 0])('rejects non-string fields from another reference kind: %p', (value) => {
  const raw = valid();
  const card = { ...raw.blocks[0].cards[0], instance_id: value };
  expect(() => parseHomeLayout({ ...raw, blocks: [{ ...raw.blocks[0], cards: [card] }] })).toThrow(
    'Mixed Home references'
  );
});
