<script setup lang="ts">
  /* Signed-out users see the sign-in page; everyone else gets the application shell. */
  import { useSession } from '@/composables/useSession';
  import LoginPage from '@/components/auth/LoginPage.vue';
  import AppShell from '@/components/shell/AppShell.vue';

  const { isAuthenticated } = useSession();
</script>

<template>
  <AppShell v-if="isAuthenticated" />
  <LoginPage v-else />
</template>
