<script setup lang="ts">
  import { onBeforeMount, h, ref, computed } from 'vue';
  import { store } from './store/index';
  import Notifications from './components/hub/alerts/Notifications.vue';
  import { useRouter } from 'vue-router';
  import NavigationBar from '@/components/controls/NavigationBar.vue';
  import Logo from '@/components/controls/Logo.vue';
  import DialogHost from './components/dialogs/DialogHost.vue';
  import PermitJoinTimer from './components/controls/PermitJoinTimer.vue';
  import { MenuBarItem } from './types/controls.type';

  const router = useRouter();
  const permitJoinDuration = 120;

  const permitJoinEnabled = ref<boolean>(false);
  const dashboardEditMode = ref<boolean>(false);
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
          command: () => router.push('/deviceDashboard'),
        },
        {
          label: 'groups',
          icon: 'pi pi-mobile',
          command: () => router.push('/'),
        },
        {
          label: 'edit',
          icon: 'pi pi-mobile',
            props: (route) => ({ editMode: route.query.editMode === 'true' }),
          command: () => {
            toggleEditMode();
            router.push({ name: 'groupdashboard', query: { editMode: dashboardEditMode.value.toString() } });
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
</style>

<template>
  <main>
    <div class="app-container">
      <div class="col-12">
        <NavigationBar :items="menuItems" />
        <PermitJoinTimer
          :duration="permitJoinDuration"
          :allow-join="permitJoinEnabled"
          @statusUpdated="permitJoinEnabled = $event" />
        <Notifications />
        <DialogHost />

        <div class="content">
          <RouterView />
        </div>
      </div>
    </div>
  </main>
</template>
