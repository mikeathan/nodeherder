<script setup lang="ts">
  import { useAuth } from '@/mixins/composables/useAuthentication';

  const { signIn } = useAuth();

  const handleLogin = async () => {
    await signIn();
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
