import {
  CALIBRATION_BLOCK_MM,
  calibrationResiduals,
  normalizeDegrees,
  solveCalibration,
} from '@/domain/spatial/calibration';
import { roomPoint } from '@/domain/spatial/workspace';

const pair = (measuredX: number, measuredY: number, roomX: number, roomY: number) => ({
  slot: 1,
  measuredX,
  measuredY,
  roomX,
  roomY,
});

test('identity and translation are recovered exactly', () => {
  expect(solveCalibration([pair(0, 1000, 0, 1000), pair(1000, 1000, 1000, 1000)])).toMatchObject({
    error: '',
    solution: { rotation: 0, x: 0, y: 0, rms: 0 },
  });
  const moved = solveCalibration([pair(0, 1000, 2000, 500), pair(1000, 1000, 3000, 500)]);
  expect(moved.solution?.rotation).toBeCloseTo(0, 6);
  expect(moved.solution?.x).toBeCloseTo(2000, 6);
  expect(moved.solution?.y).toBeCloseTo(-500, 6);
});

test('a known rotation is recovered and matches the room transform', () => {
  const sensor = { x: 500, y: 700, rotation: 90, fov: 0 };
  const points = [
    { x: 0, y: 1000 },
    { x: 1000, y: 1000 },
    { x: -500, y: 2000 },
  ];
  const solved = solveCalibration(points.map((point) => {
    const room = roomPoint(point, sensor);
    return pair(point.x, point.y, room.x, room.y);
  }));
  expect(solved.error).toBe('');
  expect(solved.solution?.rotation).toBeCloseTo(90, 6);
  expect(solved.solution?.x).toBeCloseTo(500, 6);
  expect(solved.solution?.y).toBeCloseTo(700, 6);
});

test('three noisy pairs stay solvable and report their residual', () => {
  const solved = solveCalibration([pair(0, 0, 0, 0), pair(1000, 0, 1000, 10), pair(0, 1000, -10, 1000)]);
  expect(solved.error).toBe('');
  expect(solved.solution!.rms).toBeGreaterThan(0);
  expect(solved.solution!.rms).toBeLessThan(CALIBRATION_BLOCK_MM);
  expect(solved.solution!.residuals).toHaveLength(3);
  expect(solved.solution!.worst).toBeGreaterThanOrEqual(solved.solution!.rms);
});

test('invalid and degenerate pairs are blocked with a reason', () => {
  expect(solveCalibration([pair(0, 0, 0, 0)]).error).toContain('between 2 and 5');
  expect(solveCalibration([pair(0, 0, 0, 0), pair(0, 0, 1000, 1000)]).error).toContain('collapse');
  expect(solveCalibration([pair(0, 0, 0, 0), pair(Number.NaN, 1000, 1000, 1000)]).error).toContain('finite');
  expect(solveCalibration([pair(0, 0, 0, 0), pair(1000, 1000, 5000, -5000)]).error).toContain('disagree');
});

test('angles normalise to (-180, 180]', () => {
  expect(normalizeDegrees(0)).toBe(0);
  expect(normalizeDegrees(180)).toBe(180);
  expect(normalizeDegrees(-180)).toBe(180);
  expect(normalizeDegrees(270)).toBe(-90);
  expect(normalizeDegrees(-270)).toBe(90);
  expect(calibrationResiduals([pair(0, 0, 0, 0)], { rotation: 0, x: 0, y: 0 })).toEqual([0]);
});
