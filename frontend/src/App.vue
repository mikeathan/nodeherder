<script setup lang="ts">
  /* Keep sign-in standalone until its protected destination has finished loading. */
  import { computed } from 'vue';
  import { useRoute } from 'vue-router';
  import { useSession } from '@/composables/useSession';
  import LoginPage from '@/components/auth/LoginPage.vue';
  import AppShell from '@/components/shell/AppShell.vue';

  const { isAuthenticated } = useSession();
  const route = useRoute();
  const showAppShell = computed(() => isAuthenticated.value && route.meta.requiresAuth === true);
</script>

<template>
  <AppShell v-if="showAppShell" />
  <LoginPage v-else />
</template>
