import { effectScope, nextTick, ref } from 'vue';
import { store } from '@/store';
import { useHomeLayout } from '@/composables/useHomeLayout';
import { HomeError } from '@/services/home.service';

const mockMounted: (() => void)[] = [];
const mockUnmounted: (() => void)[] = [];
const mockReplace = jest.fn();
jest.mock('vue', () => ({
  ...jest.requireActual('vue'),
  onMounted: (fn: () => void) => mockMounted.push(fn),
  onBeforeUnmount: (fn: () => void) => mockUnmounted.push(fn),
}));
jest.mock('vue-router', () => ({ useRouter: () => ({ replace: mockReplace }) }));

const flush = async () => {
  await Promise.resolve();
  await nextTick();
  await Promise.resolve();
};
let dispatch: jest.SpyInstance;
let documentEvents: EventTarget;
let windowEvents: EventTarget;
let scope: ReturnType<typeof effectScope>;
const originals = {
  add: window.addEventListener,
  remove: window.removeEventListener,
  docAdd: document.addEventListener,
  docRemove: document.removeEventListener,
};

beforeEach(() => {
  jest.useFakeTimers();
  mockMounted.length = 0;
  mockUnmounted.length = 0;
  mockReplace.mockReset();
  store.commit('home/clear');
  store.commit('home/accepted', { version: 1, revision: 1, etag: 'one', blocks: [] });
  documentEvents = new EventTarget();
  windowEvents = new EventTarget();
  window.addEventListener = windowEvents.addEventListener.bind(windowEvents);
  window.removeEventListener = windowEvents.removeEventListener.bind(windowEvents);
  document.addEventListener = documentEvents.addEventListener.bind(documentEvents);
  document.removeEventListener = documentEvents.removeEventListener.bind(documentEvents);
  dispatch = jest.spyOn(store, 'dispatch').mockResolvedValue(undefined);
  scope = effectScope();
});
afterEach(() => {
  mockUnmounted.forEach((fn) => fn());
  scope.stop();
  window.addEventListener = originals.add;
  window.removeEventListener = originals.remove;
  document.addEventListener = originals.docAdd;
  document.removeEventListener = originals.docRemove;
  jest.restoreAllMocks();
  jest.useRealTimers();
});

test('one page polls projections without per-card requests, overlaps or editing refreshes', async () => {
  const editing = ref(false);
  const home = scope.run(() => useHomeLayout({ pauseRefresh: editing }))!;
  mockMounted.forEach((fn) => fn());
  await flush();
  expect(dispatch.mock.calls.map((call) => call[0])).toEqual(['home/read', 'integrations/readInstances']);
  let finish!: () => void;
  dispatch.mockImplementation(
    () =>
      new Promise<void>((resolve) => {
        finish = resolve;
      })
  );
  jest.advanceTimersByTime(180_000);
  expect(dispatch).toHaveBeenCalledTimes(3);
  windowEvents.dispatchEvent(new Event('focus'));
  expect(dispatch).toHaveBeenCalledTimes(3);
  finish();
  await flush();
  editing.value = true;
  jest.advanceTimersByTime(60_000);
  windowEvents.dispatchEvent(new Event('focus'));
  expect(dispatch).toHaveBeenCalledTimes(3);
});

test('unmount aborts transport and releases timers, focus and visibility listeners', async () => {
  const home = scope.run(useHomeLayout)!;
  mockMounted.forEach((fn) => fn());
  await flush();
  const signal = dispatch.mock.calls[0][1] as AbortSignal;
  mockUnmounted.splice(0).forEach((fn) => fn());
  expect(signal.aborted).toBe(true);
  expect(jest.getTimerCount()).toBe(0);
  windowEvents.dispatchEvent(new Event('focus'));
  documentEvents.dispatchEvent(new Event('visibilitychange'));
  jest.advanceTimersByTime(120_000);
  expect(dispatch).toHaveBeenCalledTimes(2);
  expect(await home.save(home.layout.value)).toBe(false);
});

test('late authentication errors after session cleanup cannot sign out a replacement session', async () => {
  let reject!: (error: Error) => void;
  dispatch.mockImplementationOnce(
    () =>
      new Promise((_resolve, fail) => {
        reject = fail;
      })
  );
  const home = scope.run(useHomeLayout)!;
  const pending = home.load();
  store.commit('home/clear');
  await nextTick();
  reject(new HomeError({ code: 'unauthorized', message: 'Expired' }, 401));
  await pending;
  expect(mockReplace).not.toHaveBeenCalled();
  expect(home.error.value).toBeNull();
});

test('source failure leaves layout writable and conflicting saves remain explicit', async () => {
  dispatch.mockImplementation(async (name: string) => {
    if (name === 'integrations/readInstances') throw new Error('Unavailable');
    if (name === 'home/save') throw new HomeError({ code: 'conflict', message: 'Reload layout.' }, 409);
  });
  const home = scope.run(useHomeLayout)!;
  await home.load();
  expect(home.sourceError.value).toBe('Unavailable');
  expect(home.writable.value).toBe(true);
  expect(await home.save(home.layout.value)).toBe(false);
  expect(home.error.value).toBe('Reload layout.');
  expect(home.busy.value).toBe(false);
});
