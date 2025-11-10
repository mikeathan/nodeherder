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

  const permitJoinDuration = 120;

  const drawerWidth = ref(0);
  const isDrawerVisible = ref(false);

  const handleDrawerWidthChanged = (width: number) => {
    drawerWidth.value = width;
  };

  const toggleDrawer = () => {
    isDrawerVisible.value = !isDrawerVisible.value;
  };

  const onPermitJoinStatusUpdated = (status: boolean) => {
    isPermitJoinActive.value = status;
  };

  const { sideNavigationItems, topNavigationItems, isPermitJoinActive } = useSideNavigationItems();

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

  //  onMounted(async () => {

  //   // Restore session first to check authentication
  //   await store.dispatch('auth/restoreSession');

  //   // Only fetch hub state and connect WS if user is authenticated
  //   const isAuthenticated = store.getters['auth/isAuthenticated']();
  //   if (isAuthenticated) {
  //     fetchHubState()
  //       .then((state) => {
  //         store.dispatch('hub/init', state);
  //         store.dispatch('ws/connect');
  //       })
  //       .catch((err) => {
  //         console.error('Failed to init hub state:', err);
  //         store.commit('ws/setConnectionStatus', 'disconnected');
  //       });
  //   }
  // });
</script>

<template>
  <NavigationBar
    :style="{ marginLeft: `${drawerWidth}px` }"
    :items="topNavigationItems"
    @click="isDrawerVisible = $event" />
  <NavigationDrawer
    :items="sideNavigationItems"
    :is-expanded="isDrawerVisible"
    @toggle="toggleDrawer"
    @widthChanged="handleDrawerWidthChanged" />
  <div class="main-content" :style="{ marginLeft: `${drawerWidth}px` }">
    <PermitJoinTimer
      :duration="permitJoinDuration"
      :allow-join="isPermitJoinActive"
      @statusUpdated="onPermitJoinStatusUpdated" />
    <Notifications />
    <DialogHost />
    <RouterView />
  </div>
</template>

<style scoped>
  .main-content {
    transition: margin-left 0.5s ease;
    padding: 0 0.1rem;
  }
</style>
