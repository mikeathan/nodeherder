<script setup lang="ts">
  import { onBeforeMount, h } from 'vue';
  import { store } from './store/index';
  import Notifications from './components/hub/alerts/Notifications.vue';
  import { useRouter } from 'vue-router';
  import TimerButton from './components/controls/TimerButton.vue';
  import MenuBar from '@/components/controls/MenuBar.vue';
  import Logo from '@/components/controls/Logo.vue';
  import DialogHost from './components/dialogs/DialogHost.vue';

  const router = useRouter();
  const permitJoinDuration = 120;

  const menuItems = [
    {
      to: '/',
      label: 'dashboard',
      icon: 'pi pi-home',
      command: () => router.push('/'),
    },
    {
      to: '/',
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
      custom: true,
      template: () => h(TimerButton, { duration: permitJoinDuration }),
    },
    {
      isLogo: true,
      template: () => h(Logo),
    },
  ];

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
        <MenuBar :items="menuItems" />
        <Notifications />
        <DialogHost />

        <div class="content">
          <RouterView />
        </div>
      </div>
    </div>
  </main>
</template>
