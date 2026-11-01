import { zoneCommandPlan } from '@/domain/spatial/device-zones';
import { rectanglePoints } from '@/domain/spatial/workspace';
import type { SpatialZone } from '@/types/spatial';
import type { Expose } from '@/types/device';
const coordinate = (name: string, unit = 'mm') =>
  ({
    name,
    description: '',
    data: 0,
    category: 'config',
    values: {},
    type: 'numeric',
    unit,
    access_mode: 'readwrite',
    attributes: { min: -6000, max: 6000, step: 10 },
  }) as Expose;
const zone: SpatialZone = {
  id: 'z',
  name: 'Desk',
  kind: 'detection',
  shape: 'rectangle',
  points: rectanglePoints(-500, 1000, 1000, 1000),
  entities: { beginX: 'a', beginY: 'b', endX: 'c', endY: 'd' },
};
const fields = { a: coordinate('a'), b: coordinate('b'), c: coordinate('c'), d: coordinate('d') };
test('plans all native-unit values before submission', () => {
  expect(zoneCommandPlan(zone, fields).commands.map((c) => c.value)).toEqual([-500, 1000, 500, 2000]);
  expect(
    zoneCommandPlan(zone, { ...fields, a: { ...fields.a, unit: 'cm', attributes: { min: -600, max: 600, step: 1 } } })
      .commands[0].value
  ).toBe(-50);
});
test('rejects duplicate, missing, read-only, out-of-range and step-misaligned mappings atomically', () => {
  expect(zoneCommandPlan({ ...zone, entities: { ...zone.entities, endX: 'a' } }, fields).error).toContain('distinct');
  expect(zoneCommandPlan(zone, { ...fields, a: { ...fields.a, access_mode: 'read' } }).commands).toEqual([]);
  expect(zoneCommandPlan({ ...zone, points: rectanglePoints(-505, 1000, 1000, 1000) }, fields).error).toContain('step');
  expect(zoneCommandPlan({ ...zone, points: rectanglePoints(-7000, 1000, 1000, 1000) }, fields).commands).toEqual([]);
  expect(zoneCommandPlan({ ...zone, shape: 'polygon' }, fields).commands).toEqual([]);
  expect(
    zoneCommandPlan(zone, {
      ...fields,
      a: { ...fields.a, attributes: { ...fields.a.attributes, control_unavailable: 'Unsupported' } },
    }).commands
  ).toEqual([]);
});
