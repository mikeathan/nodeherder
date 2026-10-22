import { isHubStateSnapshot } from '@/contracts/hub';
import { createAppconfig } from '@/contracts/settings';
test('snapshot guards accept empty registries and preserve missing versus zero/false samples', () => {
  expect(isHubStateSnapshot({ devices: [], config: createAppconfig() })).toBe(true);
  expect(isHubStateSnapshot({ devices: [{ id: 'd', exposes: { temperature: { data: 0 }, contact: { data: false }, absent: { data: null } } }], config: createAppconfig() })).toBe(true);
});
test.each([null, {}, { devices: [], config: {} }, { devices: [null], config: createAppconfig() }, { devices: [{ id: 'd', exposes: { value: null } }], config: createAppconfig() }])('rejects malformed shared-state containers', (value) => {
  expect(isHubStateSnapshot(value)).toBe(false);
});
