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

test('presence cards decode with a device reference and no expose', () => {
  const raw = valid();
  raw.blocks[0].cards.push({ id: 'presence', kind: 'presence', panel: true, device_id: 'radar' } as never);
  const layout = parseHomeLayout(raw);
  expect(layout.blocks[0].cards[1]).toEqual({ id: 'presence', panel: true, kind: 'presence', device_id: 'radar' });
});

test('presence overview cards decode without any source reference', () => {
  const raw = {
    version: 1, revision: 0, etag: '', blocks: [{ id: 'b', kind: 'group', name: 'Area', cards: [
      { id: 'overview', kind: 'presence_overview', panel: true },
    ] }],
  };
  expect(parseHomeLayout(raw).blocks[0].cards[0]).toEqual({ id: 'overview', panel: true, kind: 'presence_overview' });
});

test('presence cards reject mixed device, expose and integration references', () => {
  const base = { version: 1, revision: 0, etag: '', blocks: [{ id: 'b', kind: 'group', name: 'Area', cards: [] }] };
  for (const extra of [
    { expose: 'occupied' },
    { instance_id: 'instance' },
    { resource_id: 'resource' },
    { service_id: 'service' },
  ]) {
    const raw = {
      ...base,
      blocks: [{ ...base.blocks[0], cards: [{ id: 'p', kind: 'presence', panel: true, device_id: 'radar', ...extra }] }],
    };
    expect(() => parseHomeLayout(raw)).toThrow();
  }
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
