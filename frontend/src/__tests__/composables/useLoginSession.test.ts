import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useLoginSession } from '@/composables/useLoginSession';
import { useSession } from '@/composables/useSession';
import { getAuthMode } from '@/services/auth.service';
import type { AuthModuleState } from '@/store/modules/auth/state';
import type { User } from '@/types/auth.type';

jest.mock('vue', () => ({ ...jest.requireActual('vue'), onMounted: jest.fn(), onBeforeUnmount: jest.fn() }));
jest.mock('vue-router', () => ({ useRoute: jest.fn() }));
jest.mock('@/composables/useSession', () => ({ useSession: jest.fn() }));
jest.mock('@/services/auth.service', () => ({ getAuthMode: jest.fn() }));

function deferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => { resolve = done; });
  return { promise, resolve };
}

const authenticated = ref(false);
const validation = ref<AuthModuleState['validation']>('idle');
function createSession() {
  return {
    isAuthenticated: computed(() => authenticated.value),
    user: computed(() => ({} as User)),
    validation: computed(() => validation.value),
    signingIn: ref(false), error: ref(''), signIn: jest.fn(), signOut: jest.fn(),
    completeOAuthReturn: jest.fn().mockResolvedValue(false),
    resumeSession: jest.fn().mockResolvedValue(undefined),
  } satisfies ReturnType<typeof useSession>;
}
let session: ReturnType<typeof createSession>;
let windowEvents: EventTarget;
let documentEvents: EventTarget;

beforeEach(() => {
  jest.clearAllMocks();
  authenticated.value = false;
  validation.value = 'idle';
  session = createSession();
  jest.mocked(useSession).mockReturnValue(session);
  jest.mocked(useRoute).mockReturnValue({} as ReturnType<typeof useRoute>);
  jest.mocked(getAuthMode).mockResolvedValue('google');
  windowEvents = new EventTarget();
  documentEvents = new EventTarget();
  Object.assign(window, {
    addEventListener: jest.fn(windowEvents.addEventListener.bind(windowEvents)),
    removeEventListener: jest.fn(windowEvents.removeEventListener.bind(windowEvents)),
  });
  Object.assign(document, {
    visibilityState: 'visible',
    addEventListener: jest.fn(documentEvents.addEventListener.bind(documentEvents)),
    removeEventListener: jest.fn(documentEvents.removeEventListener.bind(documentEvents)),
  });
});

function mount() { jest.mocked(onMounted).mock.calls[0][0](); }
function unmount() { jest.mocked(onBeforeUnmount).mock.calls[0][0](); }

test('checks session and mode concurrently, sharing one in-flight check', async () => {
  const mode = deferred<'local'>();
  jest.mocked(getAuthMode).mockReturnValue(mode.promise);
  const login = useLoginSession();
  const first = login.check();
  expect(login.check()).toBe(first);
  expect(session.completeOAuthReturn).toHaveBeenCalledTimes(1);
  expect(getAuthMode).toHaveBeenCalledTimes(1);
  expect(login.busy.value).toBe(true);
  mode.resolve('local');
  await first;
  expect(session.resumeSession).toHaveBeenCalledTimes(1);
  expect(login.mode.value).toBe('local');
  expect(login.unavailable.value).toBe(false);
  expect(login.busy.value).toBe(false);
});

test('OAuth completion skips resume and keeps authenticated handoff busy', async () => {
  session.completeOAuthReturn.mockResolvedValue(true);
  const login = useLoginSession();
  expect(login.statusText.value).toBe('Checking your session…');
  await login.check();
  authenticated.value = true;
  expect(session.resumeSession).not.toHaveBeenCalled();
  expect(login.busy.value).toBe(true);
  expect(login.statusText.value).toBe('Opening your home…');
});

test('mode failure offers retry and successful retry clears old errors', async () => {
  jest.mocked(getAuthMode).mockRejectedValueOnce(new Error('offline'));
  const login = useLoginSession();
  await login.check();
  expect(login.mode.value).toBeNull();
  expect(login.unavailable.value).toBe(true);
  expect(login.modeError.value).toContain('how to connect');
  expect(login.busy.value).toBe(false);
  session.error.value = 'previous error';
  await login.check();
  expect(session.error.value).toBe('');
  expect(login.modeError.value).toBe('');
  expect(login.mode.value).toBe('google');
  expect(login.unavailable.value).toBe(false);
});

test('unexpected session failure settles with an actionable error', async () => {
  session.resumeSession.mockRejectedValueOnce(new Error('unexpected rejection'));
  const login = useLoginSession();
  await login.check();
  expect(login.modeError.value).toContain('session');
  expect(login.unavailable.value).toBe(true);
  expect(login.busy.value).toBe(false);
});

test('network and focus recovery only check unavailable sessions', async () => {
  const login = useLoginSession();
  mount();
  await login.check();
  windowEvents.dispatchEvent(new Event('focus'));
  expect(getAuthMode).toHaveBeenCalledTimes(1);
  validation.value = 'unavailable';
  windowEvents.dispatchEvent(new Event('online'));
  windowEvents.dispatchEvent(new Event('focus'));
  await login.check();
  expect(getAuthMode).toHaveBeenCalledTimes(2);
});

test('visibility recovery ignores background tabs', async () => {
  const login = useLoginSession();
  mount();
  await login.check();
  validation.value = 'unavailable';
  Object.assign(document, { visibilityState: 'hidden' });
  documentEvents.dispatchEvent(new Event('visibilitychange'));
  expect(getAuthMode).toHaveBeenCalledTimes(1);
  Object.assign(document, { visibilityState: 'visible' });
  documentEvents.dispatchEvent(new Event('visibilitychange'));
  await login.check();
  expect(getAuthMode).toHaveBeenCalledTimes(2);
});

test.each([true, false])('unmount removes listeners and ignores late mode success=%s', async (success) => {
  const mode = deferred<'local'>();
  const lateResult = success ? mode.promise : mode.promise.then(() => { throw new Error('late failure'); });
  jest.mocked(getAuthMode).mockReturnValue(lateResult);
  const login = useLoginSession();
  mount();
  const pending = login.check();
  unmount();
  mode.resolve('local');
  await pending;
  expect(login.mode.value).toBeNull();
  expect(login.modeError.value).toBe('');
  expect(window.removeEventListener).toHaveBeenCalledTimes(2);
  expect(document.removeEventListener).toHaveBeenCalledTimes(1);
  validation.value = 'unavailable';
  windowEvents.dispatchEvent(new Event('online'));
  documentEvents.dispatchEvent(new Event('visibilitychange'));
  await login.check();
  expect(getAuthMode).toHaveBeenCalledTimes(1);
});
