<script setup lang="ts">
  import { ref, onMounted } from 'vue';
  import Notifications from '@/components/hub/alerts/Notifications.vue';
  import NavigationDrawer from '@/components/controls/NavigationDrawer.vue';
  import DialogHost from '@/components/dialogs/DialogHost.vue';
  import PermitJoinTimer from '@/components/controls/PermitJoinTimer.vue';
  import NavigationBar from '@/components/controls/NavigationBar.vue';
  import { fetchHubState } from '@/services/hubstate.service';
  import { store } from '@/store';
  import { useSideNavigationItems } from '@/mixins/composables/useNavigationItems';
  import { useWindowSize } from '@/mixins/composables/useWindowsSize';

  const permitJoinDuration = 120;

  const isDrawerVisible = ref(false);

  const toggleDrawer = () => {
    isDrawerVisible.value = !isDrawerVisible.value;
  };

  const onPermitJoinStatusUpdated = (status: boolean) => {
    isPermitJoinActive.value = status;
  };
 
  const { sideNavigationItems, topNavigationItems, isPermitJoinActive } = useSideNavigationItems();
  const { isMobile } = useWindowSize();

  onMounted(async () => {
    // Fast-path using persisted state: if already authed, don't block the UI.
    const authedAtMount = store.getters['auth/isAuthenticated']();

    const ensureWsConnected = () => {
      const status = store.getters['ws/getConnectionStatus'];
      if (status !== 'connected' && status !== 'connecting') {
        store.dispatch('ws/connect');
      }
    };

    const initHubIfNeeded = async () => {
      const isHubInitialized = store.getters['hub/isInitialized']();
      if (!isHubInitialized) {
        try {
          const state = await fetchHubState();
          store.dispatch('hub/init', state);
        } catch (err) {
          console.error('Failed to init hub state:', err);
          store.commit('ws/setConnectionStatus', 'disconnected');
        }
      }
    };

    if (authedAtMount) {
      // Use persisted hub state immediately; refresh in background only if needed.
      ensureWsConnected();
      initHubIfNeeded();

      // Validate session in the background without blocking initial render.
      // If validation fails, route guard/AuthWrapper will handle redirect.
      store.dispatch('auth/restoreSession');
      return;
    }

    // Not authenticated yet: validate session first (e.g., after hard refresh)
    await store.dispatch('auth/restoreSession');
    const isAuthenticated = store.getters['auth/isAuthenticated']();
    if (isAuthenticated) {
      ensureWsConnected();
      await initHubIfNeeded();
    } else {
      // Ensure WS is marked disconnected if not authenticated
      store.commit('ws/setConnectionStatus', 'disconnected');
    }
  });
</script>

<template>
  <div :class="['layout-wrapper', 'flex', isMobile ? 'mobile-layout' : 'h-screen overflow-hidden']">
    <NavigationDrawer :items="sideNavigationItems" :is-expanded="isDrawerVisible" @toggle="toggleDrawer" />

    <div class="layout-main flex flex-column flex-1 min-w-0">
      <NavigationBar :items="topNavigationItems" @click="isDrawerVisible = $event" />
      <main
        :class="['pt-0', 'pr-0', 'pl-0', 'pb-2', isMobile ? 'mobile-main' : 'flex-1 overflow-auto']"
        :style="isMobile ? {} : { overscrollBehavior: 'none' }">
        <PermitJoinTimer
          :duration="permitJoinDuration"
          :allow-join="isPermitJoinActive"
          @statusUpdated="onPermitJoinStatusUpdated" />
        <Notifications />
        <DialogHost />
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style scoped>
  .mobile-layout,
  .mobile-main {
    min-height: 100dvh;
  }
  .mobile-main :deep(> .p-card) {
    min-height: 100dvh;
    border-radius: 0 !important;
    box-shadow: none !important;
    border: none !important;
  }
</style>
