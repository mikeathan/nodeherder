import { post } from '@/contracts/api';
import { decodeHistory, historyBody } from '@/contracts/history';
import { historyKey, type HistoryRequest, type HistorySeries } from '@/domain/history/types';

type Task = {
  key: string;
  request: HistoryRequest;
  users: number;
  controller: AbortController;
  promise: Promise<HistorySeries[]>;
  resolve: (s: HistorySeries[]) => void;
  reject: (e: Error) => void;
};
const tasks = new Map<string, Task>();
const queue: Task[] = [];
let active: Task | null = null;

// One app-owned queue protects the existing rate-limited endpoint. Shared views
// release their own reference; only the final subscriber cancels the request.

async function fetchHistory(task: Task) {
  const timer = setTimeout(() => task.controller.abort(), 10000);
  try {
    const response = await post('metrics/query', historyBody(task.request), { signal: task.controller.signal });
    if (!response.ok)
      throw new Error(
        response.status === 429
          ? 'History is busy. Try again shortly.'
          : 'Could not load history. Try again or choose a shorter period.'
      );
    return decodeHistory(await response.json(), task.request);
  } catch (error) {
    if (task.controller.signal.aborted) throw new Error('History request was cancelled or timed out.');
    throw error;
  } finally {
    clearTimeout(timer);
  }
}
function complete(task: Task) {
  if (tasks.get(task.key) === task) tasks.delete(task.key);
  if (active === task) active = null;
  pump();
}
function pump() {
  if (active || !queue.length) return;
  const task = queue.shift()!;
  active = task;
  void fetchHistory(task)
    .then(task.resolve, (e: unknown) => task.reject(asError(e)))
    .finally(() => complete(task));
}
function asError(value: unknown): Error {
  return value instanceof Error ? value : new Error('Could not load history.');
}
function createTask(request: HistoryRequest): Task {
  let resolve!: Task['resolve'];
  let reject!: Task['reject'];
  const promise = new Promise<HistorySeries[]>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { key: historyKey(request), request, users: 0, controller: new AbortController(), promise, resolve, reject };
}
function release(task: Task) {
  task.users--;
  if (task.users > 0 || tasks.get(task.key) !== task) return;
  tasks.delete(task.key);
  task.controller.abort();
  const index = queue.indexOf(task);
  if (index >= 0) {
    queue.splice(index, 1);
    task.reject(new Error('History request cancelled.'));
  }
}
export function acquireHistory(request: HistoryRequest) {
  const key = historyKey(request);
  let task = tasks.get(key);
  if (!task) {
    task = createTask(request);
    if (queue.length >= 32) throw new Error('Too many history requests. Close another history view.');
    tasks.set(key, task);
    queue.push(task);
  }
  task.users++;
  const current = task;
  let released = false;
  pump();
  return {
    promise: current.promise,
    release: () => {
      if (!released) {
        released = true;
        release(current);
      }
    },
  };
}
export function resetHistoryRequests() {
  for (const task of tasks.values()) {
    task.controller.abort();
    task.reject(new Error('History session ended.'));
  }
  tasks.clear();
  queue.length = 0;
}
