<script setup lang="ts">
  const baseUrl = import.meta.env.VITE_API_BASE_URL;

  const loginWithGoogle = () => {
    window.location.href = `${baseUrl}/api/auth/login`;
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
    <Button @click="loginWithGoogle">Login with Google</Button>
  </div>
</template>
