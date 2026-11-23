import { effectScope, nextTick, ref, type Ref } from 'vue';
import { useSpatialPlayback } from '@/composables/useSpatialPlayback';

type Bounds = { from: number; to: number } | null;

function setup(bounds: Ref<Bounds>, crossing: () => { id: string; from: number } | null) {
  const scope = effectScope();
  const playback = scope.run(() => useSpatialPlayback(() => bounds.value, crossing))!;
  return { scope, playback };
}

beforeEach(() => jest.useFakeTimers());
afterEach(() => jest.useRealTimers());

test('advances, stops at the end of the recording and reports that it reached the end', async () => {
  const bounds = ref<Bounds>({ from: 0, to: 1_000 });
  const { scope, playback } = setup(bounds, () => null);
  playback.play();
  jest.advanceTimersByTime(1_100);
  await nextTick();
  expect(playback.time.value).toBe(1_000);
  expect(playback.playing.value).toBe(false);
  expect(playback.reachedEnd.value).toBe(true);
  scope.stop();
});

test('holds at the recording end and continues in the successor once it loads', async () => {
  const bounds = ref<Bounds>({ from: 0, to: 1_000 });
  const next = { id: 'next', from: 5_000 };
  const { scope, playback } = setup(bounds, () => next);
  playback.play();
  jest.advanceTimersByTime(1_100);
  await nextTick();
  expect(playback.awaitingCrossing.value).toBe(true);
  expect(playback.playing.value).toBe(true);
  expect(playback.reachedEnd.value).toBe(false);
  expect(playback.time.value).toBe(1_000);
  // The caller confirmed the successor: its bounds are now the clock's bounds.
  bounds.value = { from: 5_000, to: 6_000 };
  playback.resumeCrossing();
  expect(playback.awaitingCrossing.value).toBe(false);
  expect(playback.time.value).toBe(5_000);
  jest.advanceTimersByTime(500);
  await nextTick();
  // The interval may deliver 2–3 ticks depending on where the boundary fell, so assert that the
  // clock moved on from the successor's start rather than an exact sample.
  expect(playback.time.value).toBeGreaterThan(5_000);
  expect(playback.time.value).toBeLessThanOrEqual(5_600);
  expect(playback.playing.value).toBe(true);
  scope.stop();
});

test('pause keeps the playhead and play resumes from it', () => {
  const bounds = ref<Bounds>({ from: 0, to: 10_000 });
  const { scope, playback } = setup(bounds, () => null);
  playback.seek(2_000);
  playback.play();
  jest.advanceTimersByTime(600);
  playback.toggle();
  expect(playback.playing.value).toBe(false);
  expect(playback.time.value).toBe(2_600);
  playback.toggle();
  jest.advanceTimersByTime(400);
  expect(playback.time.value).toBe(3_000);
  scope.stop();
});

test('a failed crossing stops playback and keeps the confirmed time', () => {
  const bounds = ref<Bounds>({ from: 0, to: 1_000 });
  const { scope, playback } = setup(bounds, () => ({ id: 'next', from: 5_000 }));
  playback.play();
  jest.advanceTimersByTime(1_100);
  playback.failCrossing();
  expect(playback.playing.value).toBe(false);
  expect(playback.awaitingCrossing.value).toBe(false);
  expect(playback.time.value).toBe(1_000);
  scope.stop();
});

test('seeking during a pending crossing cancels it instead of jumping backwards later', () => {
  const bounds = ref<Bounds>({ from: 0, to: 1_000 });
  const { scope, playback } = setup(bounds, () => ({ id: 'next', from: 5_000 }));
  playback.play();
  jest.advanceTimersByTime(1_100);
  expect(playback.awaitingCrossing.value).toBe(true);
  playback.seek(200);
  playback.resumeCrossing();
  expect(playback.time.value).toBe(200);
  scope.stop();
});

test('a crossing that settles after the operator paused still moves the playhead into bounds', () => {
  const bounds = ref<Bounds>({ from: 0, to: 1_000 });
  const { scope, playback } = setup(bounds, () => ({ id: 'next', from: 5_000 }));
  playback.play();
  jest.advanceTimersByTime(1_100);
  playback.toggle();
  expect(playback.playing.value).toBe(false);
  bounds.value = { from: 5_000, to: 6_000 };
  playback.resumeCrossing();
  expect(playback.time.value).toBe(5_000);
  expect(playback.playing.value).toBe(false);
  scope.stop();
});

test('a superseded crossing is dropped so the clock cannot freeze mid-advance', () => {
  const bounds = ref<Bounds>({ from: 0, to: 1_000 });
  const { scope, playback } = setup(bounds, () => ({ id: 'next', from: 5_000 }));
  playback.play();
  jest.advanceTimersByTime(1_100);
  expect(playback.awaitingCrossing.value).toBe(true);
  // A newer selection supersedes the pending one; the caller confirms different bounds.
  playback.cancelCrossing();
  bounds.value = { from: 9_000, to: 10_000 };
  jest.advanceTimersByTime(400);
  // Without the cancel, every tick returned early on awaitingCrossing and time stayed at 1_000.
  expect(playback.time.value).toBeGreaterThan(9_000);
  expect(playback.playing.value).toBe(true);
  scope.stop();
});

test('clears the interval when the owning scope is disposed', () => {
  const bounds = ref<Bounds>({ from: 0, to: 10_000 });
  const { scope, playback } = setup(bounds, () => null);
  playback.play();
  expect(jest.getTimerCount()).toBe(1);
  scope.stop();
  expect(jest.getTimerCount()).toBe(0);
  expect(playback.playing.value).toBe(false);
});

test('will not play without bounds', () => {
  const bounds = ref<Bounds>(null);
  const { scope, playback } = setup(bounds, () => null);
  playback.play();
  expect(playback.playing.value).toBe(false);
  expect(jest.getTimerCount()).toBe(0);
  scope.stop();
});

test('play(false) keeps a playhead already at the end instead of rewinding to the start', () => {
  const bounds = ref<Bounds>({ from: 0, to: 1_000 });
  const { scope, playback } = setup(bounds, () => null);
  playback.seek(1_000);
  playback.play(false);
  // The last recorded sample of a recording sits on its `to`, so "play from this mark" used to
  // rewind to `from` and play the whole recording instead.
  expect(playback.time.value).toBe(1_000);
  expect(playback.playing.value).toBe(true);
  scope.stop();
});
