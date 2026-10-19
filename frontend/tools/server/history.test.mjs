import test from 'node:test';
import assert from 'node:assert/strict';
import { historyResponse } from './history.mjs';
const devices = [
  {
    id: 'fixture',
    exposes: {
      action: { name: 'action', type: 'enum', unit: '', data: 'press' },
      contact: { name: 'contact', type: 'binary', unit: '', data: true },
      state: { name: 'state', type: 'binary', unit: '', data: 'OFF', values: { on: 'ON', off: 'OFF' } },
      temperature: { name: 'temperature', type: 'numeric', unit: '°C', data: 20 },
    },
  },
];
const body = {
  history: true,
  deviceIds: ['fixture'],
  exposes: ['action', 'contact', 'state', 'temperature', 'missing'],
  time: { from: '2026-10-17T10:00:00Z', to: '2026-10-17T14:00:00Z' },
  limit: 10001,
};
test('mock history follows producer grouped units/raw repeated-event contract', () => {
  const result = historyResponse(body, devices);
  assert.equal(result.length, 4);
  assert.equal(result[0].values[0].value.data.length, 25);
  assert.equal(result[0].values[0].value.data[1].value, result[0].values[0].value.data[2].value);
  assert.equal(result[3].values[0].unit, '°C');
  assert.equal(typeof result[1].values[0].value.data[0].value, 'boolean');
  assert.deepEqual(
    [...new Set(result[2].values[0].value.data.map((point) => point.value))].sort(),
    ['OFF', 'ON']
  );
  assert.equal(historyResponse({ ...body, limit: 1 }, devices)[0].values[0].value.partial, true);
  assert.throws(() => historyResponse({ ...body, time: {} }, devices));
  assert.throws(() => historyResponse(body, []));
});
