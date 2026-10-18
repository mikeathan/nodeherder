import { decodeHistory, historyBody } from '@/contracts/history';
import type { HistoryRequest } from '@/domain/history/types';
const request: HistoryRequest = { deviceId: 'd', exposes: ['temperature'], from: 100, to: 200 };
function response(value: unknown = 0) {
  return [
    {
      expose: 'temperature',
      values: [
        {
          device_id: 'd',
          unit: '°C',
          value: {
            name: 'temperature',
            type: 'numeric',
            from: 100,
            to: 200,
            data: [{ timestamp: 150, value }],
            partial: false,
          },
        },
      ],
    },
  ];
}
test('decodes actual producer shape, measured zero, unknown and repeated raw events', () => {
  for (const value of [0, 1, null]) expect(decodeHistory(response(value), request)[0].data[0].value).toBe(value);
  for (const value of [true, false, 'press']) {
    const raw = response(value);
    raw[0].values[0].value.type = typeof value === 'boolean' ? 'binary' : 'enum';
    expect(decodeHistory(raw, request)[0].data[0].value).toBe(value);
  }
  const raw = response();
  raw[0].values[0].value.data = [
    { timestamp: 180, value: 4 },
    { timestamp: 140, value: 0 },
    { timestamp: 90, value: 2 },
  ];
  expect(decodeHistory(raw, request)[0].data.map((p) => p.timestamp)).toEqual([140, 180]);
  expect(historyBody(request)).toMatchObject({ history: true, deviceIds: ['d'], limit: 10001 });
});
test('rejects malformed, oversized, wrong identity, type and range responses', () => {
  for (const value of [undefined, 1, 'oops', {}, [1]]) expect(() => decodeHistory(value, request)).toThrow();
  const changes = [
    (r: ReturnType<typeof response>) => {
      r[0].values[0].device_id = 'other';
    },
    (r: ReturnType<typeof response>) => {
      r[0].values[0].value.from = 99;
    },
    (r: ReturnType<typeof response>) => {
      r[0].values[0].value.type = 'unsupported';
    },
    (r: ReturnType<typeof response>) => {
      r[0].values[0].value.data[0].timestamp = NaN;
    },
    (r: ReturnType<typeof response>) => {
      r[0].values[0].value.data[0].value = {};
    },
    (r: ReturnType<typeof response>) => {
      r[0].values[0].value.name = 'wrong';
    },
    (r: ReturnType<typeof response>) => {
      r[0].expose = 'wrong';
    },
    (r: ReturnType<typeof response>) => {
      r[0].values[0].value.data = Array(10002).fill({ timestamp: 150, value: 0 });
    },
  ];
  for (const change of changes) {
    const raw = response();
    change(raw);
    expect(() => decodeHistory(raw, request)).toThrow();
  }
  expect(() => decodeHistory(Array(33).fill(response()[0]), request)).toThrow();
});
test('marks truncation explicitly', () => {
  const raw = response();
  raw[0].values[0].value.data = Array(10001).fill({ timestamp: 150, value: 0 });
  expect(decodeHistory(raw, request)[0]).toMatchObject({ partial: true, data: expect.any(Array) });
  expect(decodeHistory(raw, request)[0].data).toHaveLength(10000);
});
