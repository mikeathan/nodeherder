<script setup lang="ts">
  /*
   * Sign-in (spec 007 US-08, FR-11, AC-25/26). OAuth redirect only; the hub owns the session
   * cookie. Handles the provider's return (/?auth=success) on mount.
   */
  import { onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useSession } from '@/composables/useSession';
  import AppLogo from '@/components/ui/AppLogo.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import GoogleIcon from '@/components/icons/GoogleIcon.vue';

  const route = useRoute();
  const { signIn, signingIn, error, completeOAuthReturn, resumeSession } = useSession();
  const year = new Date().getFullYear();

  onMounted(async () => {
    if (!(await completeOAuthReturn(route))) await resumeSession(route);
  });
</script>

<template>
  <div class="nh-login">
    <main class="nh-login-card" aria-labelledby="login-title">
      <AppLogo :size="72" class="nh-login-logo" />
      <div>
        <h1 id="login-title">Welcome home</h1>
        <p class="nh-muted">Sign in to see and control your devices.</p>
      </div>

      <button type="button" class="nh-google" :disabled="signingIn" :aria-busy="signingIn || undefined" @click="signIn">
        <span v-if="signingIn" class="nh-spin" aria-hidden="true" />
        <GoogleIcon v-else class="nh-google-ic" aria-hidden="true" />
        <span>{{ signingIn ? 'Signing in…' : 'Sign in with Google' }}</span>
      </button>

      <p v-if="error" class="nh-alert is-danger" role="alert"><UiIcon name="error" /><span>{{ error }}</span></p>
      <p class="nh-login-note"><UiIcon name="secure" />You will be sent to Google to sign in, then brought back here.</p>
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
