import { fuseSensors, type FusedSensor } from '@/domain/spatial/fusion';

const sensor = (overrides: Partial<FusedSensor> = {}): FusedSensor => ({
  id: 'a',
  name: 'Hall',
  online: true,
  occupied: false,
  targetCount: 0,
  zones: [],
  ...overrides,
});

test('occupancy is an OR and only known values decide', () => {
  expect(fuseSensors([sensor({ occupied: true }), sensor({ id: 'b', name: 'Lounge', occupied: false })]).occupied).toBe(true);
  expect(fuseSensors([sensor(), sensor({ id: 'b', name: 'Lounge' })]).occupied).toBe(false);
  expect(fuseSensors([sensor({ occupied: null }), sensor({ id: 'b', name: 'Lounge', occupied: null })]).occupied).toBeNull();
  expect(fuseSensors([]).occupied).toBeNull();
});

test('targets sum only known counts and are not people', () => {
  const result = fuseSensors([
    sensor({ targetCount: 2 }),
    sensor({ id: 'b', name: 'Lounge', targetCount: 1 }),
    sensor({ id: 'c', name: 'Office', targetCount: null, occupied: null }),
  ]);
  expect(result.targetCount).toBe(3);
  expect(result.total).toBe(3);
  expect(result.online).toBe(3);
});

test('disagreement and offline sensors are explicit', () => {
  const result = fuseSensors([
    sensor({ occupied: true, targetCount: 2 }),
    sensor({ id: 'b', name: 'Lounge', occupied: false, targetCount: 0 }),
    sensor({ id: 'c', name: 'Office', online: false, occupied: null, targetCount: null }),
  ]);
  expect(result.notes.some((note) => note.includes('occupied while others report clear'))).toBe(true);
  expect(result.notes.some((note) => note.includes('counts differ by 2'))).toBe(true);
  expect(result.offline).toEqual(['Office']);
  expect(result.notes.some((note) => note.includes('offline; they are not counted as clear'))).toBe(true);
  expect(result.online).toBe(2);
});
