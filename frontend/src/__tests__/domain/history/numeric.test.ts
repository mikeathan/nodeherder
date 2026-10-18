import { decodeHistory } from '@/contracts/history';
import { numericStats, decimate } from '@/domain/history/numeric';

test('statistics distinguish no values, zero and sparse samples', () => {
  expect(numericStats([])).toBeNull();
  expect(numericStats([{ timestamp: 0, value: null }])).toBeNull();
  expect(
    numericStats([
      { timestamp: 0, value: 0 },
      { timestamp: 1, value: 4 },
    ])
  ).toEqual({ min: 0, max: 4, avg: 2, last: 4 });
});
test('decimation preserves endpoints, gaps and a brief spike within render budget', () => {
  const data = Array.from({ length: 10000 }, (_, timestamp) => ({ timestamp, value: 0 as number | null }));
  data[5001].value = 800;
  data[7002].value = null;
  const result = decimate(data);
  expect(result.length).toBeLessThanOrEqual(1000);
  expect(result[0]).toBe(data[0]);
  expect(result[result.length - 1]).toBe(data[9999]);
  expect(result).toContain(data[5001]);
  expect(result).toContain(data[7002]);
  expect(decimate([{ timestamp: 1, value: 0 }])).toHaveLength(1);
  expect(decimate(data.map((p) => ({ ...p, value: null }))).length).toBeLessThanOrEqual(1000);
});

test('separate null runs stay separate after reducing long histories', () => {
  const source = Array.from({ length: 10000 }, (_, timestamp) => ({
    timestamp,
    value: timestamp % 200 === 100 ? null : timestamp,
  }));
  const result = decimate(source);
  expect(result.filter((p) => p.value === null)).toHaveLength(50);
  expect(result.length).toBeLessThanOrEqual(1000);
  expect(decimate(source.map((p) => ({ ...p, value: p.timestamp % 2 ? null : 1 })))).toEqual([]);
});

test('10,000-point normalization statistics and reduction stay within the proposed 100ms p95 budget', () => {
  const points = Array.from({ length: 10000 }, (_, timestamp) => ({ timestamp, value: Math.sin(timestamp / 100) }));
  const request = { deviceId: 'benchmark', exposes: ['temperature'], from: 0, to: 10000 };
  const raw = [
    {
      expose: 'temperature',
      values: [
        {
          device_id: 'benchmark',
          unit: '°C',
          value: { name: 'temperature', type: 'numeric', from: 0, to: 10000, data: points, partial: false },
        },
      ],
    },
  ];
  const timings: number[] = [];
  for (let i = 0; i < 23; i++) {
    const start = performance.now();
    const normalized = decodeHistory(raw, request)[0].data;
    numericStats(normalized);
    decimate(normalized);
    if (i >= 3) timings.push(performance.now() - start);
  }
  timings.sort((a, b) => a - b);
  const p95 = timings[18];
  console.log(
    `History domain benchmark: samples=10000, warmedRuns=20, p95=${p95.toFixed(3)}ms, payloadBytes=${new TextEncoder().encode(JSON.stringify(raw)).byteLength}, timingsMs=${JSON.stringify(timings)}`
  );
  expect(p95).toBeLessThanOrEqual(100);
});
