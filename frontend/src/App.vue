<script setup lang="ts">
  /* Keep sign-in standalone until its protected destination has finished loading. */
  import { computed, defineAsyncComponent } from 'vue';
  import { useRoute } from 'vue-router';
  import { useSession } from '@/composables/useSession';
  import LoginPage from '@/components/auth/LoginPage.vue';
  import AppShellLoading from '@/components/shell/AppShellLoading.vue';
  const AppShell = defineAsyncComponent({
    loader: () => import('@/components/shell/AppShell.vue'),
    loadingComponent: AppShellLoading,
    delay: 0,
  });

  const { isAuthenticated } = useSession();
  const route = useRoute();
  const showAppShell = computed(() => isAuthenticated.value && route.meta.requiresAuth === true);
</script>

<template>
  <AppShell v-if="showAppShell" />
  <LoginPage v-else />
</template>
