<script setup lang="ts">
  import { useAuth } from '@/mixins/composables/useAuthentication';

  const baseUrl = import.meta.env.VITE_API_BASE_URL;
  const { user, login } = useAuth();

  const handleLogin = async () => {
    const username = 'testuser';
    const res = await fetch(`${baseUrl}/api/auth/login?username=${username}`);
    if (!res.ok) {
      console.error('Login failed');
      return;
    }
    const data = await res.json();

    login(data);
  };
</script>

<style scoped>
  /* .dashboard {
  background-color: var(--p-card-background);
  height: 100vh;
  overflow: auto;
} */
</style>

<template>
  <div class="flex align-items-center pb-3 gap-1">
    <Button @click="handleLogin">Login with Google</Button>
  </div>
</template>
