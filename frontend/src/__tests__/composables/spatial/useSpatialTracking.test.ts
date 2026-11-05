import { effectScope, nextTick, ref } from 'vue';
import { store } from '@/store';
import { useSpatialTracking } from '@/composables/spatial/useSpatialTracking';
import { device, hubConfig } from '../../helpers/devices';

const mockUnmounted: (() => void)[] = [];
jest.mock('vue', () => ({
  ...jest.requireActual('vue'),
  onBeforeUnmount: (fn: () => void) => mockUnmounted.push(fn),
}));

const SPATIAL_BASE_MS = Date.parse('2026-11-03T10:00:00.000Z');
const mappings = ref([{ x: 'target_x', y: 'target_y', active: 'target_active' }]);

function frame(atMs: number, active: boolean, x = 1200, y = 2000) {
  return {
    id: 'radar',
    at: new Date(atMs).toISOString(),
    targets: [{ id: 1, active, ...(active ? { x: { value: x, unit: 'mm' }, y: { value: y, unit: 'mm' } } : {}) }],
  };
}

function setup(
  online = ref(true),
  sampled = ref<{ id: number; x: number; y: number; distance: number }[]>([])
) {
  const scope = effectScope();
  const tracking = scope.run(() =>
    useSpatialTracking(
      () => ({ id: 'radar' }),
      () => mappings.value,
      () => online.value,
      () => sampled.value
    )
  )!;
  return { scope, tracking, online, sampled };
}

beforeEach(() => {
  jest.useFakeTimers();
  jest.setSystemTime(SPATIAL_BASE_MS);
  store.dispatch('hub/init', { devices: [device('radar', { provider: 'esphome' })], config: hubConfig() });
  store.commit('ws/setConnectionStatus', 'connected');
  mappings.value = [{ x: 'target_x', y: 'target_y', active: 'target_active' }];
});
afterEach(() => {
  // Run the real unmount hooks so subscription and interval cleanup is exercised.
  mockUnmounted.splice(0).forEach((fn) => fn());
  store.commit('hub/clear');
  jest.clearAllTimers();
  jest.useRealTimers();
});

test('unmounting releases the age timer', async () => {
  const { scope } = setup();
  await nextTick();
  expect(jest.getTimerCount()).toBeGreaterThan(0);
  mockUnmounted.splice(0).forEach((fn) => fn());
  scope.stop();
  expect(jest.getTimerCount()).toBe(0);
});

test('the smoothing level reaches the filter pipeline', async () => {
  const off = setup();
  const high = setup();
  off.tracking.smoothing.value = 'off';
  high.tracking.smoothing.value = 'high';
  for (let step = 0; step < 6; step++) {
    store.commit(
      'hub/updateSpatialTargets',
      frame(SPATIAL_BASE_MS + step * 250, true, 1000 + step * 200, 2000)
    );
    jest.advanceTimersByTime(250);
    await nextTick();
  }
  const offX = off.tracking.tracks.value[0]?.x;
  const highX = high.tracking.tracks.value[0]?.x;
  expect(offX).toBeDefined();
  expect(highX).toBeDefined();
  // Raw follows the last sample exactly; the smoother lags behind it.
  expect(offX).toBeCloseTo(2000, 0);
  expect(Math.abs((offX as number) - (highX as number))).toBeGreaterThan(10);
  off.scope.stop();
  high.scope.stop();
});

test('changing the smoothing level replays recent history immediately', async () => {
  const { scope, tracking } = setup();
  for (let step = 0; step < 6; step++) {
    store.commit(
      'hub/updateSpatialTargets',
      frame(SPATIAL_BASE_MS + step * 250, true, 1000 + step * 200, 2000)
    );
    jest.advanceTimersByTime(250);
    await nextTick();
  }
  const filtered = tracking.tracks.value[0]?.x as number;
  expect(filtered).toBeLessThan(2000);

  // No new frame is needed: the level change replays the recent window, so a still target
  // still shows the effect of the control.
  tracking.smoothing.value = 'off';
  await nextTick();
  expect(tracking.tracks.value[0]?.x).toBeCloseTo(2000, 0);
  scope.stop();
});

test('prefers complete frames and converts their units', async () => {
  const { scope, tracking } = setup();
  store.commit('hub/updateSpatialTargets', frame(SPATIAL_BASE_MS, true, 120, 200));
  await nextTick();
  expect(tracking.source.value).toBe('frames');
  expect(tracking.tracks.value).toHaveLength(1);
  expect(tracking.tracks.value[0].x).toBeGreaterThan(0);
  expect(tracking.slots.value[0]).toMatchObject({ id: 1 });
  scope.stop();
});

test('a sampled coordinate cannot flip the source or reset identities', async () => {
  const sampled = ref([{ id: 1, x: 500, y: 600, distance: Math.hypot(500, 600) }]);
  const { scope, tracking } = setup(ref(true), sampled);
  store.commit('hub/updateSpatialTargets', frame(SPATIAL_BASE_MS, true));
  await nextTick();
  expect(tracking.source.value).toBe('frames');
  const label = tracking.tracks.value[0].label;

  // The ordinary observation policy still publishes a coordinate occasionally; that must not
  // replace the synchronized view or its track identity.
  sampled.value = [{ id: 1, x: 700, y: 800, distance: Math.hypot(700, 800) }];
  await nextTick();
  expect(tracking.source.value).toBe('frames');
  expect(tracking.tracks.value[0].label).toBe(label);
  scope.stop();
});

test('falls back to sampled coordinates until the first frame arrives', async () => {
  const sampled = ref([{ id: 1, x: 500, y: 600, distance: Math.hypot(500, 600) }]);
  const { scope, tracking } = setup(ref(true), sampled);
  await nextTick();
  expect(tracking.source.value).toBe('sampled');
  expect(tracking.tracks.value.map((track) => track.label)).toEqual(['Track 1']);

  store.commit('hub/updateSpatialTargets', frame(SPATIAL_BASE_MS, true));
  await nextTick();
  expect(tracking.source.value).toBe('frames');
  scope.stop();
});

test('ages a still frame by arrival time, not by the hub clock', async () => {
  const { scope, tracking } = setup();
  store.commit('hub/updateSpatialTargets', frame(SPATIAL_BASE_MS + 86_400_000, true));
  await nextTick();
  // A frame timestamped far in the future must not be treated as ancient or as fresh forever:
  // age comes from when the browser received it.
  expect(tracking.tracks.value.map((track) => track.state)).toEqual(['live']);
  jest.advanceTimersByTime(4000);
  await nextTick();
  expect(tracking.tracks.value.map((track) => track.state)).toEqual(['idle']);
  scope.stop();
});

test('shows a rejected jump as predicted and recovers when samples agree', async () => {
  const { scope, tracking } = setup();
  store.commit('hub/updateSpatialTargets', frame(SPATIAL_BASE_MS, true, 1000, 2000));
  await nextTick();
  jest.advanceTimersByTime(250);
  store.commit('hub/updateSpatialTargets', frame(SPATIAL_BASE_MS + 250, true, 6000, 2000));
  await nextTick();
  expect(tracking.tracks.value[0].state).toBe('predicted');
  expect(tracking.tracks.value[0].x).toBeLessThan(2000);

  jest.advanceTimersByTime(250);
  store.commit('hub/updateSpatialTargets', frame(SPATIAL_BASE_MS + 500, true, 6100, 2000));
  await nextTick();
  expect(tracking.tracks.value[0].state).toBe('live');
  expect(tracking.tracks.value[0].x).toBeGreaterThan(5000);
  scope.stop();
});

test('holds a cleared slot visibly and removes it within the documented window', async () => {
  const { scope, tracking } = setup();
  store.commit('hub/updateSpatialTargets', frame(SPATIAL_BASE_MS, true));
  await nextTick();
  expect(tracking.tracks.value.map((track) => track.state)).toEqual(['live']);

  jest.advanceTimersByTime(250);
  store.commit('hub/updateSpatialTargets', frame(SPATIAL_BASE_MS + 250, false));
  await nextTick();
  expect(tracking.tracks.value.map((track) => track.state)).toEqual(['held']);

  jest.advanceTimersByTime(2000);
  await nextTick();
  expect(tracking.tracks.value).toHaveLength(0);
  scope.stop();
});

test('clears the view when the device goes offline and when no mapping is saved', async () => {
  const online = ref(true);
  const { scope, tracking } = setup(online);
  store.commit('hub/updateSpatialTargets', frame(SPATIAL_BASE_MS, true));
  await nextTick();
  expect(tracking.tracks.value.length).toBe(1);

  online.value = false;
  await nextTick();
  expect(tracking.tracks.value).toHaveLength(0);
  expect(tracking.source.value).toBe('none');

  online.value = true;
  mappings.value = [{ x: '', y: '', active: '' }];
  await nextTick();
  expect(tracking.source.value).toBe('none');
  scope.stop();
});

test('ignores malformed frames without disturbing the last good position', async () => {
  const { scope, tracking } = setup();
  store.commit('hub/updateSpatialTargets', frame(SPATIAL_BASE_MS, true));
  await nextTick();
  store.commit('hub/updateSpatialTargets', { id: 'radar', at: 'bad', targets: [] });
  await nextTick();
  expect(tracking.tracks.value).toHaveLength(1);
  scope.stop();
});
