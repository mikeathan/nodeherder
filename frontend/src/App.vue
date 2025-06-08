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
  const router = useRouter();
  const permitJoinDuration = 120;

  const permitJoinEnabled = ref<boolean>(false);
  const dashboardEditMode = ref<boolean>(false);
  const drawerWidth = ref(0);
  const isDrawerMinimised = ref(true);

  const handleDrawerWidthChanged = (width: number) => {
    drawerWidth.value = width;
  };
  const toggleEditMode = () => {
    dashboardEditMode.value = !dashboardEditMode.value;
  };

  const menuItems = computed<MenuBarItem[]>(() => [
    {
      label: 'dashboards',
      icon: 'pi pi-home',
      children: [
        {
          label: 'devices',
          icon: 'pi pi-mobile',
          // command: () => router.push('/deviceDashboard'),
          command: () => {
            isDrawerMinimised.value = !isDrawerMinimised.value;
          },
          // <Button icon="pi pi-bars" @click="visible = !visible" />
        },
        {
          label: 'groups',
          icon: 'pi pi-mobile',
          command: () => router.push('/'),
        },
        {
          label: 'edit',
          icon: 'pi pi-mobile',
          command: () => {
            toggleEditMode();
            router.push({
              name: 'groupdashboard',
              params: { mode: dashboardEditMode.value ? DashboardModes.editMode : '' },
            });
          },
        },
      ],
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
      label: permitJoinEnabled.value ? 'join enabled' : 'permit Join',
      icon: 'pi pi-sitemap',
      disabled: permitJoinEnabled.value,
      command: () => {
        permitJoinEnabled.value = !permitJoinEnabled.value;
      },
    },
    {
      isLogo: true,
      template: () => h(Logo),
    },
  ]);

  onBeforeMount(() => {
    store.dispatch('ws/connect');
  });
</script>
<style scoped>
  body {
    font-family: 'Roboto', sans-serif !important;
  }

  .p-component {
    font-family: 'Roboto', sans-serif !important;
  }
  .main-content {
    transition: margin-left 0.3s ease;
    padding: 1rem;
  }
</style>

<template>
  <NavigationDrawer :items="menuItems" :is-minimised="isDrawerMinimised" @widthChanged="handleDrawerWidthChanged" />
  <div class="main-content" :style="{ marginLeft: `${drawerWidth}px` }">
    <PermitJoinTimer
      :duration="permitJoinDuration"
      :allow-join="permitJoinEnabled"
      @statusUpdated="permitJoinEnabled = $event" />
    <Notifications />
    <DialogHost />

    <RouterView />
  </div>
</template>
