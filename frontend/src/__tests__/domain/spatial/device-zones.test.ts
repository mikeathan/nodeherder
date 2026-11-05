import { serializeZonePolygon, zoneCommandPlan, zonePolygonPlan } from '@/domain/spatial/device-zones';
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

const polygonText = (name: string) =>
  ({
    name,
    description: '',
    data: '',
    category: 'config',
    values: {},
    type: 'text',
    unit: '',
    access_mode: 'readwrite',
    attributes: {},
  }) as Expose;
const polygonZone: SpatialZone = {
  id: 'p',
  name: 'Lounge',
  kind: 'detection',
  shape: 'polygon',
  points: [
    { x: 0, y: 1000 },
    { x: 1000, y: 1000 },
    { x: 500, y: 2000 },
  ],
  entities: { polygon: 'poly' },
};
const polygonFields = { poly: polygonText('poly'), coord: coordinate('coord') };

test('serialises a polygon in draw order as integer millimetres', () => {
  expect(serializeZonePolygon(polygonZone.points)).toEqual({ error: '', value: '0:1000;1000:1000;500:2000' });
  expect(zonePolygonPlan(polygonZone, polygonFields)).toEqual({
    error: '',
    commands: [{ role: 'polygon', entity: 'poly', value: '0:1000;1000:1000;500:2000' }],
  });
});

test('blocks polygons the firmware parser cannot represent', () => {
  expect(serializeZonePolygon([{ x: 0, y: 0 }, { x: 100, y: 0 }]).error).toContain('between 3 and 20');
  const many = Array.from({ length: 21 }, (_, index) => ({ x: index, y: 1000 }));
  expect(serializeZonePolygon(many).error).toContain('between 3 and 20');
  expect(
    serializeZonePolygon([{ x: 0, y: 1000 }, { x: 10.5, y: 1000 }, { x: 500, y: 2000 }]).error
  ).toContain('whole millimetres');
  expect(
    serializeZonePolygon([{ x: -6100, y: 1000 }, { x: 1000, y: 1000 }, { x: 500, y: 2000 }]).error
  ).toContain('within X');
  expect(
    serializeZonePolygon([{ x: 0, y: 1000 }, { x: 1000, y: 1000 }, { x: 500, y: 7000 }]).error
  ).toContain('within X');
});

test('a polygon needs its own text entity and accepts no rectangle roles', () => {
  expect(zonePolygonPlan(polygonZone, { ...polygonFields, poly: { ...polygonText('poly'), access_mode: 'read' } }).error).toContain('text entity');
  expect(zonePolygonPlan({ ...polygonZone, entities: { beginX: 'coord' } }, polygonFields).error).toContain('text entity');
  expect(zonePolygonPlan({ ...polygonZone, shape: 'rectangle' }, polygonFields).error).toContain('polygon zones');
  // The rectangle path still refuses polygon-shaped zones.
  expect(zoneCommandPlan(polygonZone, polygonFields).commands).toEqual([]);
});
