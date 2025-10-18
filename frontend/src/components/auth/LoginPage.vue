<script setup lang="ts">
  import { useAuth } from '@/mixins/composables/useAuthentication';
  import { ref } from 'vue';

  import GoogleIcon from '@/components/icons/GoogleIcon.vue';

  const error = ref<string>('');
  const loading = ref(false);

  const { user, signIn } = useAuth();

  const handleLogin = async () => {
    loading.value = true;
    error.value = '';
    try {
      await signIn();
    } catch (e) {
      console.error('Login failed', e);
      error.value = 'Login failed. Please try again.';
    } finally {
      loading.value = false;
    }
  };
</script>

<style scoped>
  .login-wrapper {
    height: 20vh;
    width: 100%;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
  }

  .error-message {
    color: red;
    font-size: 0.875rem;
  }

  .google-button {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;

    font-weight: 500;
    transition: background-color 0.2s, box-shadow 0.2s;
    min-width: 200px;

    background-color: #131314;
    background-image: none;
    border: 1px solid #747775;
    -webkit-border-radius: 20px;
    border-radius: 20px;
    -webkit-box-sizing: border-box;
    box-sizing: border-box;
    color: #e3e3e3;
    cursor: pointer;
    font-family: 'Roboto', arial, sans-serif;
    font-size: 14px;
    height: 40px;
    letter-spacing: 0.25px;
    outline: none;
    overflow: hidden;
    padding: 0 12px;
    position: relative;
    text-align: center;
  }

  .google-button:hover {
    box-shadow: 0 1px 2px rgba(60, 64, 67, 0.3), 0 1px 3px rgba(60, 64, 67, 0.15);
  }

  .google-button:active {
    background-color: rgba(0, 0, 0, 0.08);
  }

  .google-button:disabled {
    opacity: 0.5;
    cursor: default;
  }

  .icon {
    width: 20px;
    height: 20px;
  }
</style>

<!-- <template>
  <div class="login-wrapper flex flex-column align-items-center justify-content-center gap-3">
    <Button @click="handleLogin" :disabled="loading">
      <span v-if="!loading">Login with Google</span>
      <span v-else>Signing in...</span>
    </Button> 
    <Button @click="handleLogin" :disabled="loading" class="google-login-button flex align-items-center gap-2">
      <GoogleIcon class="google-icon" />
      <span v-if="!loading">Sign in with Google</span>
      <span v-else>Signing in...</span>
    </Button>
    <p v-if="error" class="error-message">{{ error }}</p>
  </div>
</template> -->

<template>
  <div class="login-wrapper flex flex-column align-items-center justify-content-center gap-3">
    <button class="google-button" :disabled="loading" @click="handleLogin">
      <GoogleIcon class="icon" />
      <span>{{ loading ? 'Signing in...' : 'Sign in with Google' }}</span>
    </button>
  </div>
</template>
