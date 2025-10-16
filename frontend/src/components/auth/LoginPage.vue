<script setup lang="ts">
  import { ref } from 'vue';

  const baseUrl = import.meta.env.VITE_API_BASE_URL;
  const user = ref(null);

  const login = async () => {
    const username = 'testuser';
    const res = await fetch(`${baseUrl}/api/auth/login?username=${username}`);
    const data = await res.json();

    // store JWT in localStorage or your store
    localStorage.setItem('jwt', data.token);

    user.value = data.user;
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
  <div class="flex align-items-center pb-3 gap-1">
    <Button @click="login">Login with Google</Button>
  </div>
</template>
