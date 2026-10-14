import { computed, onBeforeUnmount, onMounted, readonly, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useSession } from '@/composables/useSession';
import { getAuthMode } from '@/services/auth.service';
import type { AuthMode } from '@/services/auth.service';

/** Owns sign-in checks and recovery for the lifetime of the standalone login page. */
export function useLoginSession() {
  const route = useRoute();
  const session = useSession();
  const mode = ref<AuthMode | null>(null);
  const checking = ref(true);
  const modeError = ref('');
  const unavailable = computed(() => session.validation.value === 'unavailable' || modeError.value !== '');
  const busy = computed(() => checking.value || session.isAuthenticated.value);
  const statusText = computed(() => session.isAuthenticated.value ? 'Opening your home…' : 'Checking your session…');
  let active = true;
  let pending: Promise<void> | null = null;

  async function checkSession() {
    if (await session.completeOAuthReturn(route)) return;
    await session.resumeSession(route);
  }

  async function runCheck() {
    const [sessionResult, modeResult] = await Promise.allSettled([checkSession(), getAuthMode()]);
    if (!active) return;

    if (modeResult.status === 'fulfilled') mode.value = modeResult.value;
    else modeError.value = 'Could not check how to connect to the hub.';

    if (sessionResult.status === 'rejected') modeError.value = 'Could not check your session with the hub.';
  }

  function check(): Promise<void> {
    if (!active) return Promise.resolve();
    if (pending) return pending;

    checking.value = true;
    mode.value = null;
    modeError.value = '';
    session.error.value = '';
    pending = runCheck().finally(() => {
      pending = null;
      if (active) checking.value = false;
    });
    return pending;
  }

  function recover() {
    if (unavailable.value) void check();
  }

  function onVisibilityChange() {
    if (document.visibilityState === 'visible') recover();
  }

  onMounted(() => {
    void check();
    window.addEventListener('online', recover);
    window.addEventListener('focus', recover);
    document.addEventListener('visibilitychange', onVisibilityChange);
  });

  onBeforeUnmount(() => {
    active = false;
    window.removeEventListener('online', recover);
    window.removeEventListener('focus', recover);
    document.removeEventListener('visibilitychange', onVisibilityChange);
  });

  return {
    mode: readonly(mode),
    modeError: readonly(modeError),
    busy,
    statusText,
    unavailable,
    check,
    signIn: session.signIn,
    signingIn: readonly(session.signingIn),
    error: readonly(session.error),
  };
}
