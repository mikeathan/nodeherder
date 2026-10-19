import { acquireHistory, resetHistoryRequests } from '@/services/history.service';
import { post } from '@/contracts/api';
import type { HistoryRequest } from '@/domain/history/types';
jest.mock('@/contracts/api', () => ({ post: jest.fn() }));
const request: HistoryRequest = { deviceId: 'd', exposes: ['temperature'], from: 100, to: 200 };
const mockedPost = post as jest.MockedFunction<typeof post>;
const response = () => ({ ok: true, json: async () => [] }) as unknown as Response;
const flush = async () => {
  for (let i = 0; i < 8; i++) await Promise.resolve();
};
beforeEach(() => {
  jest.useFakeTimers();
  mockedPost.mockReset();
});
afterEach(async () => {
  resetHistoryRequests();
  await flush();
  jest.useRealTimers();
});

test('deduplicates shared subscribers, serializes different queries and releases all deadlines', async () => {
  let resolve!: (r: Response) => void;
  mockedPost
    .mockImplementationOnce(
      () =>
        new Promise((yes) => {
          resolve = yes;
        })
    )
    .mockResolvedValue(response());
  const first = acquireHistory(request);
  const same = acquireHistory({ ...request, exposes: ['temperature'] });
  const other = acquireHistory({ ...request, deviceId: 'other' });
  expect(first.promise).toBe(same.promise);
  expect(mockedPost).toHaveBeenCalledTimes(1);
  first.release();
  first.release(); // releasing twice must not cancel the other subscriber
  expect((mockedPost.mock.calls[0][2]?.signal as AbortSignal).aborted).toBe(false);
  resolve(response());
  await same.promise;
  await other.promise;
  await flush();
  expect(mockedPost).toHaveBeenCalledTimes(2);
  same.release();
  other.release();
  expect(jest.getTimerCount()).toBe(0);
});

test('queued cancellation, reset and queue bounds do not launch abandoned work', async () => {
  mockedPost.mockImplementation(
    (_url, _body, options) =>
      new Promise((_yes, no) =>
        options?.signal?.addEventListener('abort', () => no(new DOMException('aborted', 'AbortError')))
      )
  );
  const active = acquireHistory(request);
  const pending = Array.from({ length: 32 }, (_, i) => acquireHistory({ ...request, deviceId: `queued-${i}` }));
  const outcomes = pending.map((p) => p.promise.catch((e) => e));
  const ended = active.promise.catch((e) => e);
  expect(() => acquireHistory({ ...request, deviceId: 'overflow' })).toThrow('Too many');
  pending[0].release();
  resetHistoryRequests();
  await Promise.all([...outcomes, ended]);
  await flush();
  expect(mockedPost).toHaveBeenCalledTimes(1);
  expect(jest.getTimerCount()).toBe(0);
});

test('timeouts abort the request and HTTP failures are useful', async () => {
  mockedPost.mockImplementationOnce(
    (_url, _body, options) =>
      new Promise((_yes, no) =>
        options?.signal?.addEventListener('abort', () => no(new DOMException('aborted', 'AbortError')))
      )
  );
  const handle = acquireHistory(request);
  const rejected = expect(handle.promise).rejects.toThrow('cancelled or timed out');
  jest.advanceTimersByTime(10000);
  await rejected;
  await flush();
  for (const status of [429, 500]) {
    mockedPost.mockResolvedValueOnce({ ok: false, status } as Response);
    const h = acquireHistory(request);
    await expect(h.promise).rejects.toThrow(status === 429 ? 'busy' : 'Could not');
    await flush();
  }
  mockedPost.mockRejectedValueOnce('bad transport');
  await expect(acquireHistory(request).promise).rejects.toThrow('Could not');
});

test('last subscriber cancellation aborts active network work', async () => {
  mockedPost.mockImplementationOnce(
    (_url, _body, options) =>
      new Promise((_yes, no) => options?.signal?.addEventListener('abort', () => no(new Error('aborted'))))
  );
  const handle = acquireHistory(request);
  const rejected = expect(handle.promise).rejects.toThrow('cancelled or timed out');
  handle.release();
  await rejected;
  await flush();
  expect(jest.getTimerCount()).toBe(0);
});
