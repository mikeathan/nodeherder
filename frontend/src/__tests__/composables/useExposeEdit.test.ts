import { effectScope, nextTick } from 'vue';
import { store } from '@/store';
import { useExposeEdit } from '@/composables/useExposeEdit';
import { device, expose, hubConfig } from '../helpers/devices';
import { captureWsFrames } from '../helpers/ws-capture';

beforeEach(() => {
  jest.useFakeTimers();
});
afterEach(() => {
  store.commit('hub/clear');
  jest.runOnlyPendingTimers();
  jest.useRealTimers();
});
test('preview, cancellation, confirmation, refusal and timeout never overwrite hub truth', async () => {
  const light = device('edit-light', {}, [
    expose('brightness', { data: 50, attributes: { min: 0, max: 254 }, access_mode: 'readwrite' }),
  ]);
  store.dispatch('hub/init', { devices: [light], config: hubConfig() });
  store.commit('ws/setConnectionStatus', 'connected');
  const capture = captureWsFrames();
  const scope = effectScope();
  const find = () => store.getters['hub/findDevice'](light.id);
  const edit = scope.run(() => useExposeEdit(find, () => find().exposes.brightness))!;
  edit.dragging.value = true;
  edit.preview(100);
  store.commit('hub/updateDevice', { id: light.id, data: { brightness: 75 } });
  await nextTick();
  expect(edit.value.value).toBe(100);
  edit.cancel();
  expect(edit.value.value).toBe(75);
  expect(capture.frames).toHaveLength(0);
  edit.preview(120);
  edit.commitNumber(120);
  await nextTick();
  expect(edit.pending.value).toBe(true);
  expect(find().exposes.brightness.data).toBe(75);
  jest.advanceTimersByTime(5000);
  await nextTick();
  expect(edit.value.value).toBe(75);
  edit.preview(130);
  edit.commitNumber(130);
  await nextTick();
  store.commit('hub/updateDevice', { id: light.id, data: { brightness: 130 } });
  await nextTick();
  expect(edit.pending.value).toBe(false);
  expect(edit.value.value).toBe(130);
  store.commit('ws/setConnectionStatus', 'disconnected');
  edit.preview(140);
  edit.commitNumber(140);
  expect(edit.value.value).toBe(130);
  edit.commitNumber(NaN);
  expect(capture.frames).toHaveLength(2);
  scope.stop();
  capture.restore();
});

test('missing readings have a usable native position without displaying a measured zero', () => {
  const light = device('unknown-light', {}, [
    expose('brightness', { data: null, attributes: { min: 2, max: 254 }, access_mode: 'readwrite' }),
  ]);
  const scope = effectScope();
  const edit = scope.run(() =>
    useExposeEdit(
      () => light,
      () => light.exposes.brightness
    )
  )!;
  expect(edit.value.value).toBe(2);
  expect(edit.displayValue.value).toBeNull();
  edit.preview(100);
  expect(edit.displayValue.value).toBe(100);
  edit.cancel();
  expect(edit.displayValue.value).toBeNull();
  scope.stop();
});
