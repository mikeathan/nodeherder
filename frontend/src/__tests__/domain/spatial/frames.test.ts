import { axisMillimetres, frameSamples, parseSpatialFrame } from '@/domain/spatial/frames';

const valid = {
  id: 'sensor',
  at: '2026-11-03T10:00:00.000Z',
  targets: [
    { id: 1, active: true, x: { value: 1200, unit: 'mm' }, y: { value: 3400, unit: 'mm' } },
    { id: 2, active: false },
  ],
};

test('rejects malformed envelopes instead of committing them', () => {
  expect(parseSpatialFrame(null)).toBeNull();
  expect(parseSpatialFrame({ targets: [] })).toBeNull();
  expect(parseSpatialFrame({ id: 'sensor', at: 'not-a-date', targets: [] })).toBeNull();
  expect(parseSpatialFrame({ id: 'sensor', at: valid.at, targets: 'nope' })).toBeNull();
});

test('drops unusable targets and axes while keeping the rest of the frame', () => {
  const frame = parseSpatialFrame({
    id: 'sensor',
    at: valid.at,
    targets: [
      { id: 4, active: true, x: { value: 1, unit: 'mm' }, y: { value: 1, unit: 'mm' } },
      { id: 1, active: true, x: { value: 10, unit: 'furlong' }, y: { value: 20, unit: 'mm' } },
      { id: 1, active: true, x: { value: 30, unit: 'mm' }, y: { value: 40, unit: 'mm' } },
      { id: 2, active: 'yes' },
    ],
  });
  expect(frame?.targets).toEqual([{ id: 1, active: true, y: { value: 20, unit: 'mm' } }]);
});

test('converts declared units to millimetres and keeps missing axes null', () => {
  expect(axisMillimetres({ value: 120, unit: 'cm' })).toBe(1200);
  expect(axisMillimetres({ value: 1.5, unit: 'm' })).toBe(1500);
  expect(axisMillimetres({ value: 10, unit: 'ft' })).toBeNull();
  expect(axisMillimetres(undefined)).toBeNull();
  expect(axisMillimetres({ value: Number.NaN, unit: 'mm' })).toBeNull();

  const frame = parseSpatialFrame(valid);
  expect(frame).not.toBeNull();
  const samples = frameSamples(frame!);
  expect(samples[0]).toEqual({ slot: 1, active: true, x: 1200, y: 3400 });
  expect(samples[1]).toEqual({ slot: 2, active: false, x: null, y: null });
});

test('caps slots and keeps the first reading for a repeated slot', () => {
  const frame = parseSpatialFrame({
    id: 'sensor',
    at: valid.at,
    targets: [
      { id: 1, active: true, x: { value: 1, unit: 'mm' }, y: { value: 1, unit: 'mm' } },
      { id: 1, active: false },
      { id: 2, active: false },
      { id: 3, active: false },
      { id: 4, active: false },
    ],
  });
  expect(frame?.targets.map((target) => target.id)).toEqual([1, 2, 3]);
  expect(frame?.targets[0].active).toBe(true);
});
