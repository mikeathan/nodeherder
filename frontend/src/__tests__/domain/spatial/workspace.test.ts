import {
  containsPoint,
  copyWorkspace,
  createWorkspace,
  rectanglePoints,
  resizeRectangle,
  polygonBounds,
  roomScreenPoint,
  sensorPoint,
  validPolygon,
  validateWorkspace,
  zonePreview,
} from '@/domain/spatial/workspace';
import type { SpatialZone } from '@/types/spatial';
const zone: SpatialZone = {
  id: 'desk',
  name: 'Desk',
  kind: 'detection',
  shape: 'rectangle',
  points: rectanglePoints(-500, 1000, 1000, 1000),
};
test('rejects malformed polygons and workspace bounds without rejecting empty legacy settings', () => {
  expect(validateWorkspace(createWorkspace())).toBe('');
  expect(
    validPolygon([
      { x: 0, y: 0 },
      { x: 100, y: 100 },
      { x: 0, y: 100 },
      { x: 100, y: 0 },
    ])
  ).toBe(false);
  expect(validPolygon(rectanglePoints(0, 0, 0, 100))).toBe(false);
  const w = createWorkspace();
  w.zones = [zone];
  expect(validateWorkspace(w)).toBe('');
  w.sensor.x = NaN;
  expect(validateWorkspace(w)).not.toBe('');
  w.sensor.x = 0;
  w.zones = Array.from({ length: 5 }, (_, i) => ({ ...zone, id: String(i) }));
  expect(validateWorkspace(w)).toContain('slots');
});
test('room transforms invert correctly and keep persisted drafts independent', () => {
  const sensor = { x: 1200, y: -300, rotation: 90, fov: 100 };
  const p = { x: 400, y: 2000 };
  const screen = roomScreenPoint(p, sensor, 270);
  const restored = sensorPoint(screen, sensor, 270);
  expect(restored.x).toBeCloseTo(p.x);
  expect(restored.y).toBeCloseTo(p.y);
  const w = createWorkspace();
  w.zones = [zone];
  const copied = copyWorkspace(w);
  copied.zones[0].points[0].x = 999;
  expect(w.zones[0].points[0].x).toBe(-500);
});
test('inclusive occupancy previews respect exclusions and never call missing targets clear', () => {
  const target = { id: 1, x: 0, y: 1500, distance: 1500 };
  expect(containsPoint(zone.points, { x: -500, y: 1000 })).toBe(true);
  expect(zonePreview(zone, [zone], [target], true)).toBe('Target inside');
  expect(zonePreview(zone, [zone, { ...zone, id: 'mask', kind: 'exclusion' }], [target], true)).toBe(
    'No target inside'
  );
  expect(zonePreview(zone, [zone], [], true)).toContain('Unknown');
  expect(zonePreview(zone, [zone], [target], false)).toContain('Unknown');
});

test('rectangle resizing retains a fixed anchor across both axes and shares numeric bounds', () => {
  const anchor = { x: 1000, y: 1000 };
  expect(polygonBounds(resizeRectangle(anchor, { x: 1200, y: 1200 }))).toEqual({
    x: 1000,
    y: 1000,
    width: 200,
    height: 200,
  });
  expect(polygonBounds(resizeRectangle(anchor, { x: 1400, y: 1400 }))).toEqual({
    x: 1000,
    y: 1000,
    width: 400,
    height: 400,
  });
  expect(polygonBounds(resizeRectangle(anchor, { x: -500, y: 1500 }))).toEqual({
    x: -500,
    y: 1000,
    width: 1500,
    height: 500,
  });
  expect(anchor).toEqual({ x: 1000, y: 1000 });
});
