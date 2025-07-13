<script setup lang="ts">
  import { onBeforeMount, h, ref, computed } from 'vue';
  import { store } from './store/index';
  import Notifications from './components/hub/alerts/Notifications.vue';
  import { useRouter } from 'vue-router';
  import NavigationDrawer from '@/components/controls/NavigationDrawer.vue';
  import Logo from '@/components/controls/Logo.vue';
  import DialogHost from './components/dialogs/DialogHost.vue';
  import PermitJoinTimer from './components/controls/PermitJoinTimer.vue';
  import { MenuBarItem } from './types/controls.type';
  import { DashboardModes } from '@/types/controls.type';
  import NavigationBar from './components/controls/NavigationBar.vue';
  const router = useRouter();
  const permitJoinDuration = 120;

  const isPermitJoinActive = ref<boolean>(false);
  const dashboardEditMode = ref<boolean>(false);
  const drawerWidth = ref(0);
  const isDrawerVisible = ref(false);

  const handleDrawerWidthChanged = (width: number) => {
    drawerWidth.value = width;
  };

  const toggleDrawer = () => {
    isDrawerVisible.value = !isDrawerVisible.value;
  }
  const toggleEditMode = () => {
    dashboardEditMode.value = !dashboardEditMode.value;
  };
  const onPermitJoinStatusUpdated = (status: boolean) => {
    isPermitJoinActive.value = status;
  };

  const startPermitJoinTimer = () => {
    if (!isPermitJoinActive.value) {
      isPermitJoinActive.value = true;
    }
  };

  const topNavigattionItems = computed<MenuBarItem[]>(() => [
    {
      isLogo: true,
      template: () => h(Logo),
    },
    {
      icon: 'pi pi-sitemap',
      get disabled() {
        return isPermitJoinActive.value;
      },
      command: () => startPermitJoinTimer(),
    },
    {
      icon: 'pi pi-cog',
      command: () => {
        toggleEditMode();
        router.push({
          name: 'groupdashboard',
          params: { mode: dashboardEditMode.value ? DashboardModes.editMode : '' },
        });
      },
    },
  ]);

  const sideNavigationItems = computed<MenuBarItem[]>(() => [
    {
      label: 'groups',
      icon: 'pi pi-home',
      command: () => router.push('/'),
    },
    {
      label: 'devices',
      icon: 'pi pi-mobile',
      command: () => router.push('/deviceDashboard'),
    },
    {
      to: '/devicelist',
      label: 'device list',
      icon: 'pi pi-list',
      command: () => router.push('/devicelist'),
    },
    {
      to: '/viewer',
      label: 'automations',
      icon: 'pi pi-objects-column',
      command: () => router.push('/viewer'),
    },
    {
      to: '/consoleviewer',
      label: 'console',
      icon: 'pi pi-code',
      command: () => router.push('/consoleviewer'),
    },
    {
      to: '/settings',
      label: 'settings',
      icon: 'pi pi-cog',
      command: () => router.push('/settings'),
    },
    {
      label: 'permit Join',
      icon: 'pi pi-sitemap',
      get disabled() {
        return isPermitJoinActive.value;
      },
      command: () => startPermitJoinTimer(),
    },
  ]);

  onBeforeMount(() => {
    store.dispatch('ws/connect');
  });
</script>

<template>
  <NavigationBar
    :style="{ marginLeft: `${drawerWidth}px` }"
    :items="topNavigattionItems"
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
