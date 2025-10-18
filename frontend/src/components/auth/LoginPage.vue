<script setup lang="ts">
  import { useAuth } from '@/mixins/composables/useAuthentication';
  import { ref } from 'vue';

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
</style>

<template>
  <div class="login-wrapper flex flex-column align-items-center justify-content-center gap-3">
    <Button @click="handleLogin" :disabled="loading">
      <span v-if="!loading">Login with Google</span>
      <span v-else>Signing in...</span>
    </Button>

    <p v-if="error" class="error-message">{{ error }}</p>
  </div>
</template>
