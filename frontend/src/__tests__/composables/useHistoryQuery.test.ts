import { effectScope, nextTick, ref } from 'vue';
import { store } from '@/store';
import { useHistoryQuery } from '@/composables/useHistoryQuery';
import { acquireHistory } from '@/services/history.service';
import type { HistoryRequest } from '@/domain/history/types';
jest.mock('@/services/history.service', () => ({ acquireHistory: jest.fn(), resetHistoryRequests: jest.fn() }));
const request: HistoryRequest = { deviceId: 'd', exposes: ['temperature'], from: 100, to: 200 };
const acquire = acquireHistory as jest.MockedFunction<typeof acquireHistory>;
const flush = async () => {
  await Promise.resolve();
  await nextTick();
};
beforeEach(() => {
  jest.useFakeTimers();
  acquire.mockReset();
  store.commit('metrics/clear');
});
afterEach(() => {
  jest.useRealTimers();
});

test('debounces changes, releases old subscribers and rejects late results', async () => {
  let resolveFirst!: (result: []) => void;
  const release = jest.fn();
  acquire
    .mockReturnValueOnce({
      promise: new Promise((yes) => {
        resolveFirst = yes;
      }),
      release,
    })
    .mockReturnValue({ promise: Promise.resolve([]), release: jest.fn() });
  const source = ref<HistoryRequest | null>(request);
  const scope = effectScope();
  const query = scope.run(() => useHistoryQuery(source))!;
  expect(query.entry.value?.status).toBe('loading');
  jest.advanceTimersByTime(300);
  await flush();
  source.value = { ...request, from: 101 };
  await nextTick();
  expect(release).toHaveBeenCalledTimes(1);
  resolveFirst([]);
  await flush();
  expect(query.entry.value?.status).toBe('loading');
  jest.advanceTimersByTime(300);
  await flush();
  expect(query.entry.value?.status).toBe('ready');
  scope.stop();
  expect(jest.getTimerCount()).toBe(0);
});

test('reset/logout prevents abandoned completion from repopulating the cache', async () => {
  let resolve!: (value: []) => void;
  acquire.mockReturnValue({
    promise: new Promise((yes) => {
      resolve = yes;
    }),
    release: jest.fn(),
  });
  const scope = effectScope();
  const q = scope.run(() => useHistoryQuery(() => request))!;
  jest.advanceTimersByTime(300);
  await flush();
  store.commit('metrics/clear');
  resolve([]);
  await flush();
  expect(q.entry.value).toBeUndefined();
  scope.stop();
});

test('errors and null identity settle clearly and cache entries stay bounded', async () => {
  acquire.mockReturnValue({ promise: Promise.reject(new Error('Unavailable')), release: jest.fn() });
  const source = ref<HistoryRequest | null>(request);
  const scope = effectScope();
  const q = scope.run(() => useHistoryQuery(source))!;
  jest.advanceTimersByTime(300);
  await flush();
  expect(q.entry.value?.error).toBe('Unavailable');
  q.refresh();
  await nextTick();
  jest.advanceTimersByTime(300);
  await flush();
  source.value = null;
  await nextTick();
  jest.advanceTimersByTime(300);
  await flush();
  expect(q.entry.value).toBeUndefined();
  for (let i = 0; i < 20; i++)
    store.commit('metrics/historySet', { key: String(i), entry: { status: 'ready', series: [], error: '' } });
  expect(store.getters['metrics/history']('0')).toBeUndefined();
  expect(store.getters['metrics/history']('19')).toBeDefined();
  scope.stop();
});
