import { ref } from 'vue';
import { useConnectionDetail } from '@/composables/integrations/useConnectionDetail';
import type { ConnectionInstance, ConnectionSchedule } from '@/types/integrations.type';

let mockMounted: (() => Promise<void>) | undefined;
let mockScope: any;
let mockQueue: ConnectionInstance[] = [];

jest.mock('vue', () => ({
  ...jest.requireActual('vue'),
  onMounted: (callback: () => Promise<void>) => {
    mockMounted = callback;
  },
  onBeforeUnmount: jest.fn(),
}));
jest.mock('vue-router', () => ({
  useRoute: () => ({ params: { id: 'one' }, hash: '', query: {} }),
}));
jest.mock('@/composables/useConnections', () => ({ useConnections: () => mockScope }));
jest.mock('@/composables/useConfirm', () => ({ confirm: jest.fn() }));

const mockDispatch = jest.fn();

function schedule(intervalSeconds: number, enabled = true): ConnectionSchedule {
  return {
    enabled,
    interval_seconds: intervalSeconds,
    next_due_at: null,
    last_attempt_at: null,
    last_success_at: null,
    last_failure_at: null,
    consecutive_failures: 0,
  };
}

function instance(intervalSeconds: number): ConnectionInstance {
  return {
    id: 'one',
    name: 'Bins',
    provider_key: 'waste_collection/whenisbins',
    revision: 1,
    etag: 'etag',
    enabled: true,
    status: 'ready',
    postcode: 'XX1 1XX',
    credential_slots: [],
    credential_configured: false,
    runs: [],
    resources: [],
    schedule: schedule(intervalSeconds),
    last_error: null,
    updated_at: '2026-10-29T12:00:00Z',
  };
}

function configure(cached: ConnectionInstance, queue: ConnectionInstance[]) {
  mockQueue = [...queue];
  mockScope = {
    api: {
      definition: jest.fn().mockResolvedValue({ document: {} }),
      updateSchedule: jest.fn().mockImplementation(async (_id, _etag, _enabled, seconds: number) => {
        const saved = instance(seconds);
        saved.etag = 'etag-saved';
        saved.revision = 2;
        mockQueue.unshift(saved);
        return saved;
      }),
    },
    isActive: () => true,
    state: ref({ current: cached, instances: [cached], generation: 1 }),
    store: { dispatch: mockDispatch },
    router: { push: jest.fn(), replace: jest.fn() },
    busy: ref(false),
    error: ref(null),
    perform: async (action: () => Promise<unknown>) => action(),
    observe: jest.fn(),
    checkOperation: jest.fn(),
    signal: new AbortController().signal,
    commit: jest.fn(),
  };
  mockDispatch.mockImplementation(async () => {
    const value = mockQueue.shift();
    if (!value) throw new Error('no queued instance');
    mockScope.state.value = { current: value, instances: [value], generation: 1 };
    return value;
  });
  return useConnectionDetail();
}

const flush = async () => {
  for (let i = 0; i < 8; i++) await Promise.resolve();
};

beforeEach(() => {
  jest.useFakeTimers();
  jest.clearAllMocks();
  mockMounted = undefined;
  Object.defineProperty(globalThis, 'window', {
    value: { addEventListener: jest.fn(), removeEventListener: jest.fn() },
    configurable: true,
  });
  Object.defineProperty(globalThis, 'document', { value: { hidden: false }, configurable: true });
});

afterEach(() => {
  jest.clearAllTimers();
  jest.useRealTimers();
});

test('focus and minute reads reconcile a pristine schedule editor', async () => {
  const cached = instance(432_000);
  cached.etag = 'etag-a';
  const fresh = instance(3_600);
  fresh.etag = 'etag-b';
  const detail = configure(cached, [cached, fresh]);
  await mockMounted?.();
  expect(detail.scheduleValue.value).toBe(5);
  expect(detail.scheduleUnit.value).toBe('days');
  jest.advanceTimersByTime(60_000);
  await flush();
  expect(detail.scheduleValue.value).toBe(1);
  expect(detail.scheduleUnit.value).toBe('hours');
});

test('a later read preserves unsaved schedule edits', async () => {
  const cached = instance(432_000);
  cached.etag = 'etag-a';
  const fresh = instance(3_600);
  fresh.etag = 'etag-b';
  const detail = configure(cached, [cached, fresh]);
  await mockMounted?.();
  detail.scheduleValue.value = 30;
  detail.scheduleUnit.value = 'minutes';
  jest.advanceTimersByTime(60_000);
  await flush();
  expect(detail.scheduleValue.value).toBe(30);
  expect(detail.scheduleUnit.value).toBe('minutes');
});

test('unrelated instance revisions rebase the schedule draft', async () => {
  const cached = instance(432_000);
  cached.etag = 'etag-a';
  const fresh = instance(432_000);
  fresh.etag = 'etag-b';
  fresh.revision = 2;
  const detail = configure(cached, [cached, fresh]);
  await mockMounted?.();
  detail.scheduleValue.value = 30;
  detail.scheduleUnit.value = 'minutes';
  jest.advanceTimersByTime(60_000);
  await flush();
  await detail.saveSchedule();
  expect(mockScope.api.updateSchedule).toHaveBeenCalledWith('one', 'etag-b', true, 1_800);
});

test('a changed schedule keeps the draft base revision and conflicts on save', async () => {
  const cached = instance(432_000);
  cached.etag = 'etag-a';
  const changed = instance(3_600);
  changed.etag = 'etag-b';
  changed.revision = 2;
  const detail = configure(cached, [cached, changed]);
  await mockMounted?.();
  detail.scheduleValue.value = 30;
  detail.scheduleUnit.value = 'minutes';
  jest.advanceTimersByTime(60_000);
  await flush();
  await detail.saveSchedule();
  expect(mockScope.api.updateSchedule).toHaveBeenCalledWith('one', 'etag-a', true, 1_800);
});
