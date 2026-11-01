import { parseDeviceConfigSaveResult } from '@/contracts/device-settings';
const valid = { requestId: '00000000-0000-0000-0000-000000000001', deviceId: 'radar', success: true };
test('validates reply identities and prevents malformed replies confirming settings', () => {
  expect(parseDeviceConfigSaveResult(valid)).toEqual(valid);
  for (const value of [
    null,
    {},
    { ...valid, requestId: 'bad' },
    { ...valid, deviceId: '' },
    { ...valid, success: 'yes' },
    { ...valid, error: 'contradictory' },
    { ...valid, success: false, error: 'x'.repeat(513) },
  ])
    expect(parseDeviceConfigSaveResult(value)).toBeNull();
  expect(parseDeviceConfigSaveResult({ ...valid, success: false, error: 'Unavailable' })).toEqual({
    ...valid,
    success: false,
    error: 'Unavailable',
  });
});
