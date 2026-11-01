import { effectScope, nextTick, ref, watchEffect } from 'vue';
import { store } from '@/store';
import { useSpatialView } from '@/composables/spatial/useSpatialView';
import { useWorkspaceEditor } from '@/composables/spatial/useWorkspaceEditor';
import { device, expose, hubConfig } from '../../helpers/devices';
import { captureWsFrames } from '../../helpers/ws-capture';
const mockUnmounted: (() => void)[] = [];
jest.mock('vue', () => ({ ...jest.requireActual('vue'), onBeforeUnmount: (fn: () => void) => mockUnmounted.push(fn) }));
let capture: ReturnType<typeof captureWsFrames>;
let scope: ReturnType<typeof effectScope>;
function setup() {
  const current = ref(device('radar', { provider: 'esphome' }));
  scope = effectScope();
  const view = scope.run(() => useSpatialView(() => current.value))!;
  const editor = scope.run(() => useWorkspaceEditor(view.workspace))!;
  return { current, view, editor };
}
function reply(success = true) {
  const frame = capture.frames.at(-1)!.payload as { requestId: string; id: string };
  store.commit('ws/setDeviceConfigSaveResult', {
    requestId: frame.requestId,
    deviceId: frame.id,
    success,
    ...(!success ? { error: 'Save rejected' } : {}),
  });
}
beforeEach(() => {
  jest.useFakeTimers();
  store.dispatch('hub/init', { devices: [device('radar'), device('other')], config: hubConfig() });
  store.commit('ws/setConnectionStatus', 'connected');
  capture = captureWsFrames();
});
afterEach(() => {
  mockUnmounted.splice(0).forEach((fn) => fn());
  scope?.stop();
  store.commit('hub/clear');
  capture.restore();
  jest.clearAllTimers();
  jest.useRealTimers();
});
test('failure retains the editable draft and Undo, and retry commits only after confirmation', async () => {
  const { view, editor } = setup();
  editor.addZone('detection');
  const workspace = view.workspace.value;
  const first = view.saveDisplay();
  await nextTick();
  expect(view.saving.value).toBe(true);
  expect(view.canSave.value).toBe(false);
  expect(editor.undoStack.value).toHaveLength(1);
  await view.saveDisplay();
  expect(capture.frames).toHaveLength(1);
  reply(false);
  await first;
  await nextTick();
  expect(view.saveError.value).toBe('Save rejected');
  expect(view.workspace.value).toBe(workspace);
  expect(editor.undoStack.value).toHaveLength(1);
  expect(store.getters['hub/hasDeviceConfigOverride']('radar')).toBe(false);
  const retry = view.saveDisplay();
  reply();
  await retry;
  await nextTick();
  expect(view.saveError.value).toBe('');
  expect(view.saveNotice.value).toContain('saved');
  expect(view.workspace.value).toBe(workspace);
  expect(editor.undoStack.value).toHaveLength(1);
  expect(store.getters['hub/findDeviceSetting']('radar').spatial.workspace.zones).toHaveLength(1);
  const reloaded = JSON.parse(JSON.stringify(store.getters['hub/findDeviceSetting']('radar')));
  reloaded.spatial.workspace.zones = reloaded.spatial.workspace.zones.map((zone: Record<string, unknown>) =>
    Object.fromEntries(Object.entries(zone).reverse())
  );
  store.commit('hub/setDeviceConfigOverride', reloaded);
  await nextTick();
  expect(view.workspace.value).toBe(workspace);
  expect(editor.undoStack.value).toHaveLength(1);
});
test('edits made while saving survive an older acknowledgement and remote settings refresh', async () => {
  const { view, editor } = setup();
  editor.addZone('detection');
  const pending = view.saveDisplay();
  editor.addObject('furniture', 'Desk');
  reply();
  await pending;
  await nextTick();
  expect(view.workspace.value.objects).toHaveLength(1);
  expect(store.getters['hub/findDeviceSetting']('radar').spatial.workspace.objects).toHaveLength(0);
  const config = hubConfig();
  store.commit('hub/setAppConfig', config);
  await nextTick();
  expect(view.workspace.value.objects).toHaveLength(1);
  expect(editor.undoStack.value).toHaveLength(2);
  editor.undo();
  expect(view.workspace.value.objects).toHaveLength(0);
});
test('timeout keeps recovery history and explains unknown persistence outcome', async () => {
  const { view, editor } = setup();
  editor.addZone('entry');
  const pending = view.saveDisplay();
  jest.advanceTimersByTime(10000);
  await pending;
  expect(view.saveError.value).toContain('may have reached the hub');
  expect(editor.undoStack.value).toHaveLength(1);
  expect(view.canSave.value).toBe(true);
});
test('switching devices cancels old save ownership and ignores its late result', async () => {
  const { current, view, editor } = setup();
  editor.addZone('detection');
  const pending = view.saveDisplay();
  current.value = device('other', { provider: 'esphome' });
  await nextTick();
  reply();
  await pending;
  await nextTick();
  expect(view.workspace.value.zones).toHaveLength(0);
  expect(view.saveNotice.value).toBe('');
  expect(view.saving.value).toBe(false);
  expect(store.getters['hub/hasDeviceConfigOverride']('radar')).toBe(false);
});
test('unmount aborts pending saves and releases owned timers and subscriptions', async () => {
  const { view, editor } = setup();
  editor.addZone('detection');
  const pending = view.saveDisplay();
  mockUnmounted.splice(0).forEach((fn) => fn());
  scope.stop();
  reply();
  await pending;
  expect(store.getters['hub/hasDeviceConfigOverride']('radar')).toBe(false);
  expect(jest.getTimerCount()).toBe(0);
});

test('editor warning and label callbacks ignore live reports but react to inventory changes', async () => {
  const { current, view } = setup();
  const targetX = expose('target2_x', { unit: 'mm', attributes: { label: 'Target 2 X' } as never });
  current.value = { ...current.value, exposes: { target2_x: targetX } };
  view.setMapping(1, 'x', targetX.name);
  let renders = 0;
  let warning = '';
  let label = '';
  scope.run(() =>
    watchEffect(() => {
      renders++;
      warning = view.mappingWarning(view.mappings.value[1], 1);
      label = view.selectedLabel(targetX.name);
    })
  );
  await nextTick();
  const before = renders;
  current.value = { ...current.value, exposes: { target2_x: { ...targetX, data: 200 } } };
  await nextTick();
  expect(renders).toBe(before);
  expect(warning).toBe('');
  expect(label).toBe('Target 2 X');
  current.value = {
    ...current.value,
    exposes: { target2_x: { ...targetX, attributes: { label: 'Target 2 Y' } as never } },
  };
  await nextTick();
  expect(renders).toBeGreaterThan(before);
  expect(warning).toContain('does not match');
  expect(label).toBe('Target 2 Y');
});

test('unmapping saved Target 2 survives refresh, failed save, retry and reopening without shifting rows', async () => {
  const mappings = [
    { x: 'x1', y: 'y1', active: 'a1' },
    { x: 'x2', y: 'y2', active: 'a2' },
    { x: '', y: '', active: '' },
  ];
  const saved = {
    ...store.getters['hub/findDeviceSetting']('radar'),
    id: 'radar',
    spatial: { mappings, scale: 6000, orientation: 0 },
  };
  store.commit('hub/setDeviceConfigOverride', saved);
  const { current, view } = setup();
  current.value = device('radar', { provider: 'esphome' }, [
    expose('x1', { unit: 'mm' }), expose('y1', { unit: 'mm' }),
    expose('a1', { type: 'binary' as never }),
    expose('x2', { unit: 'mm' }), expose('y2', { unit: 'mm' }),
    expose('a2', { type: 'binary' as never }),
  ]);
  view.setMapping(1, 'x', '');
  await nextTick();
  await view.saveDisplay();
  expect(view.mappingError.value).toContain('Target 2');
  expect(capture.frames).toHaveLength(0);
  view.setMapping(1, 'y', '');
  view.setMapping(1, 'active', '');
  store.commit('hub/setDeviceConfigOverride', JSON.parse(JSON.stringify(saved)));
  await nextTick();
  expect(view.mappings.value[1]).toEqual({ x: '', y: '', active: '' });
  const failed = view.saveDisplay();
  reply(false);
  await failed;
  expect(view.mappings.value[1]).toEqual({ x: '', y: '', active: '' });
  expect(store.getters['hub/findDeviceSetting']('radar').spatial.mappings[1]).toEqual(mappings[1]);
  const retry = view.saveDisplay();
  reply();
  await retry;
  await nextTick();
  const reopened = scope.run(() => useSpatialView(() => current.value))!;
  expect(reopened.mappings.value).toEqual([mappings[0], { x: '', y: '', active: '' }, mappings[2]]);
});
