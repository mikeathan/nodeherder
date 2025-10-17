<script setup lang="ts">
  import { store } from '@/store';
  import { computed, ref } from 'vue';

  const baseUrl = import.meta.env.VITE_API_BASE_URL;

  const user = computed(() => store.getters['auth/user']());
  const login = async () => {
    const username = 'testuser';
    const res = await fetch(`${baseUrl}/api/auth/login?username=${username}`);
    const data = await res.json();

    store.dispatch('auth/loginUser', { user: data.user, token: data.token });
    console.log('Logged in with mock token', data.token);
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
  User{{ user }}
  <div class="flex align-items-center pb-3 gap-1">
    <Button @click="login">Login with Google</Button>
  </div>
</template>
