<script setup lang="ts">
  import { h, ref, computed, onMounted } from 'vue';
  import Notifications from './components/hub/alerts/Notifications.vue';
  import { useRouter } from 'vue-router';
  import NavigationDrawer from '@/components/controls/NavigationDrawer.vue';
  import Logo from '@/components/controls/Logo.vue';
  import DialogHost from './components/dialogs/DialogHost.vue';
  import PermitJoinTimer from './components/controls/PermitJoinTimer.vue';
  import NavigationBar from './components/controls/NavigationBar.vue';
  import { MenuBarItem } from '@/types/controls.type';
  import { fetchHubState } from '@/services/hubstate.service';
  import { store } from '@/store';
  import { goTo } from '@/router/navigation';
  import { RouteName } from '@/types/router';
  import { useSideNavigationItems } from '@/mixins/composables/useNavigationItems';

  const router = useRouter();
  const permitJoinDuration = 120;

  //   const isPermitJoinActive = ref<boolean>(false);
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

  //   const startPermitJoinTimer = () => {
  //     if (!isPermitJoinActive.value) {
  //       isPermitJoinActive.value = true;
  //     }
  //   };

  const { sideNavigationItems, topNavigationItems, isPermitJoinActive } = useSideNavigationItems();
  //   const topNavigationItems = computed<MenuBarItem[]>(() => [
  //     {
  //       isLogo: true,
  //       template: () => h(Logo),
  //       command: () => {},
  //     },
  //   ]);

  //   const sideNavigationItems = computed<MenuBarItem[]>(() => [
  //     {
  //       label: 'groups',
  //       icon: 'pi pi-home',
  //       command: () => goTo(RouteName.GroupDashboard),
  //     },
  //     {
  //       label: 'devices',
  //       icon: 'pi pi-mobile',
  //       command: () => goTo(RouteName.Devices),
  //     },
  //     {
  //       to: '/devicelist',
  //       label: 'device list',
  //       icon: 'pi pi-list',
  //       command: () => goTo(RouteName.DeviceList),
  //     },
  //     {
  //       to: '/viewer',
  //       label: 'automations',
  //       icon: 'pi pi-objects-column',
  //       command: () => goTo(RouteName.Viewer),
  //     },
  //     {
  //       to: '/consoleviewer',
  //       label: 'console',
  //       icon: 'pi pi-code',
  //       command: () => goTo(RouteName.ConsoleViewer),
  //     },
  //     {
  //       to: '/settings',
  //       label: 'settings',
  //       icon: 'pi pi-cog',
  //       command: () => goTo(RouteName.Settings),
  //     },
  //     {
  //       label: 'permit Join',
  //       icon: 'pi pi-sitemap',
  //       get disabled() {
  //         return isPermitJoinActive.value;
  //       },
  //       command: () => startPermitJoinTimer(),
  //     },
  //   ]);

  onMounted(() => {
    fetchHubState()
      .then((state) => {
        store.dispatch('hub/init', state);
        store.dispatch('ws/connect');
      })
      .catch((err) => {
        console.error('Failed to init hub state:', err);
        store.commit('ws/setConnectionStatus', 'disconnected');
      });
    store.dispatch('auth/restoreSession');
  });
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
  body {
    font-family: 'Roboto', sans-serif !important;
  }

  .p-component {
    font-family: 'Roboto', sans-serif !important;
  }
  .main-content {
    transition: margin-left 0.5s ease;
    padding: 0 0.1rem;
  }
</style>
