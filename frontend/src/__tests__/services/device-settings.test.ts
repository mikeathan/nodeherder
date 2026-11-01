import { store } from '@/store';
import { saveConfirmedDeviceSettings } from '@/services/device-settings';
import { captureWsFrames } from '../helpers/ws-capture';
import { device, hubConfig } from '../helpers/devices';
import { createDeviceConfigOverride } from '@/contracts/settings';
let capture: ReturnType<typeof captureWsFrames>;
const config = () => ({ ...createDeviceConfigOverride('radar'), metricsEnabled: true });
function reply(patch: Record<string, unknown> = {}) {
  const frame = capture.frames.at(-1)!.payload as { requestId: string };
  store.commit('ws/setDeviceConfigSaveResult', {
    requestId: frame.requestId,
    deviceId: 'radar',
    success: true,
    ...patch,
  });
}
beforeEach(() => {
  jest.useFakeTimers();
  store.dispatch('hub/init', { devices: [device('radar')], config: hubConfig() });
  store.commit('ws/setConnectionStatus', 'connected');
  capture = captureWsFrames();
});
afterEach(() => {
  store.commit('hub/clear');
  capture.restore();
  jest.clearAllTimers();
  jest.useRealTimers();
});
test('publishes only the matching persisted reply and snapshots the submitted value', async () => {
  const input = config();
  const pending = saveConfirmedDeviceSettings(input, new AbortController().signal);
  input.metricsEnabled = false;
  expect(store.getters['hub/hasDeviceConfigOverride']('radar')).toBe(false);
  reply({ requestId: '00000000-0000-0000-0000-000000000000' });
  reply({ deviceId: 'another' });
  reply({ success: 'yes' });
  expect(store.getters['hub/hasDeviceConfigOverride']('radar')).toBe(false);
  reply();
  await pending;
  expect(store.getters['hub/findDeviceSetting']('radar').metricsEnabled).toBe(true);
  expect(jest.getTimerCount()).toBe(0);
});
test('a rejected save leaves confirmed Vuex settings unchanged', async () => {
  const pending = saveConfirmedDeviceSettings(config(), new AbortController().signal);
  const rejected = expect(pending).rejects.toThrow('Storage unavailable');
  reply({ success: false, error: 'Storage unavailable' });
  await rejected;
  expect(store.getters['hub/hasDeviceConfigOverride']('radar')).toBe(false);
  expect(jest.getTimerCount()).toBe(0);
});
test('timeout reports unknown outcome, releases work and ignores late acknowledgement', async () => {
  const pending = saveConfirmedDeviceSettings(config(), new AbortController().signal);
  const rejected = expect(pending).rejects.toThrow('It may have reached the hub');
  jest.advanceTimersByTime(10000);
  await rejected;
  reply();
  expect(store.getters['hub/hasDeviceConfigOverride']('radar')).toBe(false);
  expect(jest.getTimerCount()).toBe(0);
});
test.each(['abort', 'disconnect', 'reset'])(
  '%s cancels ownership and ignores late acknowledgements',
  async (reason) => {
    const controller = new AbortController();
    const pending = saveConfirmedDeviceSettings(config(), controller.signal);
    const rejected = expect(pending).rejects.toThrow();
    if (reason === 'abort') controller.abort();
    if (reason === 'disconnect') store.commit('ws/setConnectionStatus', 'disconnected');
    if (reason === 'reset') store.commit('hub/clear');
    await rejected;
    reply();
    expect(store.getters['hub/hasDeviceConfigOverride']('radar')).toBe(false);
    expect(jest.getTimerCount()).toBe(0);
  }
);
