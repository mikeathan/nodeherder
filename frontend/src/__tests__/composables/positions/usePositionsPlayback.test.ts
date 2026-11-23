import { effectScope, nextTick, ref } from 'vue';
import type { SpatialSession } from '@/services/spatial-history.service';
import { usePositionsPlayback } from '@/composables/positions/usePositionsPlayback';

function session(id: string, from: number, to: number): SpatialSession {
  return { id, from, to, points: 1, slots: [1], capped: false, samples: [from] };
}

async function flush(): Promise<void> {
  await nextTick();
  await Promise.resolve();
  await nextTick();
}

beforeEach(() => jest.useFakeTimers());
afterEach(() => jest.useRealTimers());

function setup(options: {
  sessions: SpatialSession[];
  confirmedId: string;
  load: (id: string) => Promise<boolean>;
}) {
  const scope = effectScope();
  const sessions = ref(options.sessions);
  const confirmed = ref(options.confirmedId);
  const supersededAt = ref(0);
  const state = scope.run(() =>
    usePositionsPlayback({ sessions, confirmed, load: options.load, supersededAt })
  )!;
  return { scope, state, sessions, confirmed, supersededAt };
}

test('a failed crossing stops the clock and clears the pending flag', async () => {
  const { scope, state, confirmed } = setup({
    sessions: [session('a', 0, 1_000), session('b', 5_000, 6_000)],
    confirmedId: 'a',
    load: async (id) => id !== 'b',
  });
  await state.play('a');
  jest.advanceTimersByTime(1_200);
  await flush();
  // The regression: returning on a failed load without settling the crossing left every later
  // tick early-returning, with playing stuck true and Pause→Play unable to recover.
  expect(state.awaitingCrossing.value).toBe(false);
  expect(state.playing.value).toBe(false);
  expect(state.time.value).toBe(1_000);
  expect(confirmed.value).toBe('a');
  scope.stop();
});

test('a successful crossing continues into the successor', async () => {
  const { scope, state, confirmed } = setup({
    sessions: [session('a', 0, 1_000), session('b', 5_000, 6_000)],
    confirmedId: 'a',
    load: async (id) => {
      confirmed.value = id;
      return true;
    },
  });
  await state.play('a');
  jest.advanceTimersByTime(1_200);
  await flush();
  expect(confirmed.value).toBe('b');
  expect(state.awaitingCrossing.value).toBe(false);
  expect(state.time.value).toBe(5_000);
  expect(state.playing.value).toBe(true);
  scope.stop();
});

test('select confirms a recording and places the playhead at its start, paused', async () => {
  const { scope, state, confirmed } = setup({
    sessions: [session('a', 0, 1_000), session('b', 5_000, 6_000)],
    confirmedId: 'a',
    load: async (id) => {
      confirmed.value = id;
      return true;
    },
  });
  const loaded = await state.select('b');
  await flush();
  expect(loaded).toBe(true);
  expect(confirmed.value).toBe('b');
  expect(state.time.value).toBe(5_000);
  expect(state.playing.value).toBe(false);
  scope.stop();
});

test('a crossed gap is announced briefly and then yields the status line', async () => {
  const { scope, state, confirmed } = setup({
    sessions: [session('a', 0, 1_000), session('b', 500_000, 501_000)],
    confirmedId: 'a',
    load: async (id) => {
      confirmed.value = id;
      return true;
    },
  });
  await state.play('a');
  jest.advanceTimersByTime(1_200);
  await flush();
  expect(state.skippedGap.value).toEqual({ from: 1_000, to: 500_000 });
  jest.advanceTimersByTime(5_000);
  await flush();
  // Left set, the gap masked every idle/pending/end message for the rest of the recording.
  expect(state.skippedGap.value).toBeNull();
  scope.stop();
});

test('play from the recording’s final sample starts there, not at its beginning', async () => {
  const { scope, state, confirmed } = setup({
    sessions: [session('a', 0, 1_000)],
    confirmedId: 'a',
    load: async (id) => {
      confirmed.value = id;
      return true;
    },
  });
  // The last recorded sample sits exactly on `to`; activating its mark must play from that moment.
  const started = await state.play('a', 1_000);
  await flush();
  expect(started).toBe(true);
  expect(state.time.value).toBe(1_000);
  expect(state.playing.value).toBe(true);
  scope.stop();
});

test('a superseded crossing reply cannot stop a clock started on another recording', async () => {
  let releaseCrossing: (loaded: boolean) => void = () => {};
  const crossingLoad = new Promise<boolean>((resolve) => {
    releaseCrossing = resolve;
  });
  const { scope, state, confirmed } = setup({
    sessions: [session('a', 0, 1_000), session('b', 5_000, 6_000), session('c', 9_000, 10_000)],
    confirmedId: 'a',
    load: async (id) => {
      // The crossing into `b` stays pending; every other selection resolves immediately.
      if (id === 'b') return crossingLoad;
      confirmed.value = id;
      return true;
    },
  });
  await state.play('a');
  jest.advanceTimersByTime(1_200);
  await flush();
  expect(state.awaitingCrossing.value).toBe(true);

  // The operator switches to `c` and presses Play while `b`'s reply is still in flight.
  await state.play('c');
  await flush();
  expect(confirmed.value).toBe('c');
  expect(state.playing.value).toBe(true);

  // `b`'s stale reply reports failure; without the guard it called failCrossing() and stopped `c`.
  releaseCrossing(false);
  await flush();
  expect(state.playing.value).toBe(true);
  expect(confirmed.value).toBe('c');
  scope.stop();
});
