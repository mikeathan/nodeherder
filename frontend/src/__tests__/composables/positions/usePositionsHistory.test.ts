import { effectScope, nextTick } from 'vue';
import type { SpatialPoint, SpatialSession } from '@/services/spatial-history.service';
import { usePositionsHistory } from '@/composables/positions/usePositionsHistory';

jest.mock('@/services/spatial-history.service', () => ({
  spatialHistoryApi: {
    sessions: jest.fn(),
    points: jest.fn(),
    heatmap: jest.fn(),
  },
}));

import { spatialHistoryApi } from '@/services/spatial-history.service';

const sessions = spatialHistoryApi.sessions as jest.Mock;
const points = spatialHistoryApi.points as jest.Mock;
const heatmap = spatialHistoryApi.heatmap as jest.Mock;

function session(id: string, from: number, to: number): SpatialSession {
  return { id, from, to, points: 1, slots: [1], capped: false, samples: [from] };
}

function point(at: number): SpatialPoint {
  return { at, slot: 1, x: 0, y: 0 };
}

async function flush(): Promise<void> {
  await nextTick();
  await Promise.resolve();
  await nextTick();
  await Promise.resolve();
  await nextTick();
}

beforeEach(() => {
  jest.clearAllMocks();
  heatmap.mockResolvedValue([]);
});

test('loads the filter and confirms the newest recording with its points', async () => {
  sessions.mockResolvedValue([session('a', 0, 1_000), session('b', 2_000, 3_000)]);
  points.mockResolvedValue([point(2_500)]);
  const scope = effectScope();
  const history = scope.run(() => usePositionsHistory(() => 'dev'))!;
  await flush();
  expect(history.sessions.value.map((entry) => entry.id)).toEqual(['a', 'b']);
  expect(history.selected.value).toBe('b');
  expect(history.confirmed.value).toBe('b');
  expect(history.points.value).toEqual([point(2_500)]);
  scope.stop();
});

test('a superseded points reply cannot confirm and is counted for playback', async () => {
  sessions.mockResolvedValue([]);
  const scope = effectScope();
  const history = scope.run(() => usePositionsHistory(() => 'dev'))!;
  await flush();

  let resolveFirst!: (value: SpatialPoint[]) => void;
  points
    .mockImplementationOnce(() => new Promise<SpatialPoint[]>((resolve) => (resolveFirst = resolve)))
    .mockResolvedValueOnce([point(9)]);

  const first = history.loadPoints('a');
  const second = history.loadPoints('b');
  await second;
  resolveFirst([]);
  await first;
  await flush();

  expect(history.confirmed.value).toBe('b');
  expect(history.supersededAt.value).toBe(1);
  scope.stop();
});

test('a failed points load keeps the confirmed recording and reports the error', async () => {
  sessions.mockResolvedValue([]);
  const scope = effectScope();
  const history = scope.run(() => usePositionsHistory(() => 'dev'))!;
  await flush();

  history.confirmed.value = 'a';
  history.selected.value = 'b';
  points.mockRejectedValue(new Error('Spatial history could not be loaded.'));
  const loaded = await history.loadPoints('b');
  await flush();

  expect(loaded).toBe(false);
  expect(history.confirmed.value).toBe('a');
  expect(history.selected.value).toBe('a');
  expect(history.pointsError.value).toBe('Spatial history could not be loaded.');
  scope.stop();
});
