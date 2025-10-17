<script setup lang="ts">
  import router from '@/router';
  import { store } from '@/store';
  import { computed, ref } from 'vue';

  const baseUrl = import.meta.env.VITE_API_BASE_URL;

  const user = computed(() => store.getters['auth/user']());
  const login = async () => {
    const username = 'testuser';
    const res = await fetch(`${baseUrl}/api/auth/login?username=${username}`);
    if (!res.ok) {
      console.error('Login failed');
      return;
    }
    const data = await res.json();

    store.dispatch('auth/loginUser', data);

    TODO;
    // need to come up with better solution for redirect after login
    router.push({ name: 'groupdashboard' });
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
    <Button @click="login">Login with Google</Button>
  </div>
</template>
