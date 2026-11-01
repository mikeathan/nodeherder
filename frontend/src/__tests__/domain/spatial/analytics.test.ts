import { activityByHour, appendFrame, heatCells, targetTrails } from '@/domain/spatial/analytics';
const target = { id: 1, x: -1, y: 400, distance: 400 };
test('bounds session samples, copies target values and excludes expired frames', () => {
  const original = [{ at: 0, targets: [target] }];
  const next = appendFrame(original, [target], 3600001);
  expect(next).toHaveLength(1);
  next[0].targets[0].x = 100;
  expect(target.x).toBe(-1);
  expect(
    appendFrame(
      Array.from({ length: 4000 }, (_, i) => ({ at: i, targets: [] })),
      [],
      5000
    )
  ).toHaveLength(3600);
});
test('heat cells preserve negative coordinates and do not invent empty/offline measurements', () => {
  const frames = [
    { at: 1000, targets: [target] },
    { at: 2000, targets: [] },
    { at: 3000, targets: [target] },
  ];
  expect(heatCells(frames, 1500)).toEqual([{ x: -250, y: 250, count: 1, intensity: 1 }]);
  expect(heatCells([], 0)).toEqual([]);
  expect(activityByHour(frames, 0)[0].count).toBe(2);
});
test('trail paths break on target loss and timing gaps', () => {
  const frames = [
    { at: 1000, targets: [target] },
    { at: 2000, targets: [target] },
    { at: 3000, targets: [] },
    { at: 4000, targets: [target] },
    { at: 7000, targets: [target] },
  ];
  expect(targetTrails(frames, 7000)).toEqual([{ id: 1, targets: [target, target] }]);
});
