<script setup lang="ts">
  /*
   * Sign-in (spec 008): the hub selects local or Google access and owns the cookie.
   * Check an existing session before presenting sign-in; retry unavailable checks.
   */
  import { computed, onMounted, onBeforeUnmount, ref } from 'vue';
  import { useRoute } from 'vue-router';
  import { useSession } from '@/composables/useSession';
  import AppLogo from '@/components/ui/AppLogo.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import GoogleIcon from '@/components/icons/GoogleIcon.vue';
  import { getAuthMode } from '@/services/auth.service';
  import type { AuthMode } from '@/services/auth.service';

  const route = useRoute();
  const { signIn, signingIn, error, validation, completeOAuthReturn, resumeSession } = useSession();
  const mode = ref<AuthMode | null>(null);
  const checking = ref(true);
  const modeError = ref('');
  const unavailable = computed(() => validation.value === 'unavailable' || !!modeError.value);
  let active = true;
  let started = false;

  async function check() {
    if (checking.value && started) return;
    started = true;
    checking.value = true;
    mode.value = null;
    modeError.value = '';
    error.value = '';
    const sessionCheck = async () => {
      if (!(await completeOAuthReturn(route))) await resumeSession(route);
    };
    await Promise.all([
      sessionCheck(),
      getAuthMode().then((value) => { if (active) mode.value = value; }).catch(() => {
        if (active) modeError.value = 'Could not check how to connect to the hub.';
      }),
    ]);
    if (active) checking.value = false;
  }
  const recover = () => { if (unavailable.value) void check(); };
  const visible = () => { if (document.visibilityState === 'visible') recover(); };
  const year = new Date().getFullYear();

  onMounted(() => {
    void check();
    window.addEventListener('online', recover);
    window.addEventListener('focus', recover);
    document.addEventListener('visibilitychange', visible);
  });
  onBeforeUnmount(() => {
    active = false;
    window.removeEventListener('online', recover);
    window.removeEventListener('focus', recover);
    document.removeEventListener('visibilitychange', visible);
  });
</script>

<template>
  <div class="nh-login">
    <main class="nh-login-card" aria-labelledby="login-title">
      <AppLogo :size="72" class="nh-login-logo" />
      <div>
        <h1 id="login-title">Welcome home</h1>
        <p class="nh-muted">{{ mode === 'local' ? 'Connect to your hub on the local network.' : 'Connect to see and control your devices.' }}</p>
      </div>

      <p v-if="checking" role="status">Checking your session…</p>
      <template v-else-if="unavailable">
        <p class="nh-alert is-danger" role="alert">
          <UiIcon name="offline" /><span>{{ modeError || 'Could not reach the hub to check your session.' }} Check your connection and retry.</span>
        </p>
        <button type="button" class="nh-google" @click="check">Retry connection</button>
      </template>
      <template v-else>
        <button type="button" class="nh-google" :disabled="signingIn" :aria-busy="signingIn || undefined" @click="signIn">
          <span v-if="signingIn" class="nh-spin" aria-hidden="true" />
          <UiIcon v-else-if="mode === 'local'" name="home" />
          <GoogleIcon v-else class="nh-google-ic" aria-hidden="true" />
          <span>{{ signingIn ? 'Connecting…' : mode === 'local' ? 'Continue on local network' : 'Sign in with Google' }}</span>
        </button>
        <p v-if="error" class="nh-alert is-danger" role="alert"><UiIcon name="error" /><span>{{ error }}</span></p>
        <p v-if="mode === 'local'" class="nh-login-note"><UiIcon name="home" />Local access uses the shared Offline User session. No Google account is needed.</p>
        <p v-else class="nh-login-note"><UiIcon name="secure" />You will be sent to Google to sign in, then brought back here.</p>
      </template>
    </main>
    <footer class="nh-login-foot">© {{ year }} NodeHerder</footer>
  </div>
</template>

<style scoped>
  .nh-login {
    min-height: 100dvh;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 1.5rem;
    padding: 1.5rem 1rem;
    background:
      radial-gradient(50rem 30rem at 0% 0%, var(--nh-accent-soft), transparent 70%),
      radial-gradient(40rem 30rem at 100% 100%, color-mix(in srgb, var(--dk-lamp) 14%, transparent), transparent 70%),
      var(--nh-bg);
  }
  .nh-login-card {
    width: min(25rem, 100%);
    display: flex;
    flex-direction: column;
    gap: 1.25rem;
    padding: 2rem;
    background: var(--nh-surface);
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    box-shadow: var(--nh-shadow-2);
  }
  .nh-login-logo :deep(.nh-logo-word) {
    font-size: 1.5rem;
  }
  h1 {
    font-size: 1.45rem;
    margin-bottom: 0.25rem;
  }
  p {
    margin: 0;
  }
  .nh-google {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.75rem;
    width: 100%;
    min-height: 3rem;
    border-radius: 99px;
    border: var(--nh-border-w) solid var(--nh-border-strong);
    background: var(--nh-surface);
    color: var(--nh-text);
    font-weight: 600;
    cursor: pointer;
    transition:
      background var(--nh-motion),
      box-shadow var(--nh-motion);
  }
  .nh-google:hover:not(:disabled) {
    background: var(--nh-surface-2);
    box-shadow: var(--nh-shadow-1);
  }
  .nh-google:disabled {
    opacity: 0.6;
    cursor: progress;
  }
  .nh-google-ic {
    width: 1.25rem;
    height: 1.25rem;
  }
  .nh-login-note {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    font-size: 0.8rem;
    color: var(--nh-text-muted);
  }
  .nh-login-note .nh-ic {
    color: var(--nh-ok);
  }
  .nh-login-foot {
    font-size: 0.8rem;
    color: var(--nh-text-muted);
  }
  @media (max-width: 420px) {
    .nh-login-card {
      padding: 1.5rem 1.25rem;
    }
  }
</style>
