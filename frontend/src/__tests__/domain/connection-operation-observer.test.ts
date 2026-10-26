import { createOperationObserver } from '@/domain/connection-operation-observer';

const flush = async () => {
  for (let i = 0; i < 8; i++) await Promise.resolve();
};
afterEach(() => jest.useRealTimers());

function fixture() {
  const read = jest.fn().mockResolvedValue({ status: 'running' });
  const refresh = jest.fn().mockResolvedValue(undefined);
  const onError = jest.fn();
  return { read, refresh, onError, observer: createOperationObserver({ read, refresh, onError }) };
}

test('replacing an in-flight observation revokes its refresh and timer ownership', async () => {
  jest.useFakeTimers();
  const f = fixture();
  let complete!: (value: { status: string }) => void;
  f.read.mockImplementationOnce(
    () =>
      new Promise((resolve) => {
        complete = resolve;
      })
  );
  f.observer.start('old');
  const oldSignal = f.read.mock.calls[0][1] as AbortSignal;
  f.observer.start('new');
  await flush();
  complete({ status: 'running' });
  await flush();
  expect(oldSignal.aborted).toBe(true);
  expect(f.refresh).toHaveBeenCalledTimes(1);
  expect(jest.getTimerCount()).toBe(1);
  f.observer.stop();
  expect(jest.getTimerCount()).toBe(0);
});

test('stop during refresh cannot schedule a replacement timer or report an old error', async () => {
  jest.useFakeTimers();
  const f = fixture();
  let fail!: (error: Error) => void;
  f.refresh.mockImplementationOnce(
    () =>
      new Promise((_resolve, reject) => {
        fail = reject;
      })
  );
  f.observer.start('one');
  await flush();
  f.observer.stop();
  fail(new Error('late failure'));
  await flush();
  expect(jest.getTimerCount()).toBe(0);
  expect(f.onError).not.toHaveBeenCalled();
});

test('terminal progress and the 30-second deadline both stop local observation', async () => {
  jest.useFakeTimers();
  const f = fixture();
  f.read.mockResolvedValueOnce({ status: 'pending' });
  f.observer.start('pending');
  await flush();
  expect(jest.getTimerCount()).toBe(0);
  f.observer.start('running');
  await flush();
  await jest.advanceTimersByTimeAsync(31_000);
  expect(jest.getTimerCount()).toBe(0);
});
