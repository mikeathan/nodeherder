<script setup lang="ts">
  import { onBeforeMount, h, ref } from 'vue';
  import { store } from './store/index';
  import Notifications from './components/hub/alerts/Notifications.vue';
  import { useRouter } from 'vue-router';
  import TimerButton from './components/controls/TimerButton.vue';
  import NavigationBar from '@/components/controls/NavigationBar.vue';
  import Logo from '@/components/controls/Logo.vue';
  import DialogHost from './components/dialogs/DialogHost.vue';
  import { Button } from 'primevue';
import TimerPanel from './components/controls/TimerPanel.vue';

  const router = useRouter();
  const permitJoinDuration = 120;

  const permitJoinLabel = ref<string>('');
  const permitJoinEnabled = ref<boolean>(false);
  function enablePermitJoin() {
  }

  const menuItems = [
    {
      to: '/',
      label: 'group dashboard',
      icon: 'pi pi-home',
      children: [
        {
          to: '/',
          label: 'device dashboard',
          icon: 'pi pi-mobile',
          command: () => router.push('/devicedashboard'),
        },
      ],
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
      label: 'permit join',
      icon: 'pi pi-sitemap',
       command: () => { permitJoinEnabled.value = true; },
    },
    // {
    //   custom: true,
    //   template: () => h(TimerButton, { duration: permitJoinDuration }),
    // },
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
        <NavigationBar :items="menuItems" />
        <TimerPanel :duration="permitJoinDuration" :show="permitJoinEnabled"/> or click event ?
        <Notifications />
        <DialogHost />

        <div class="content">
          <RouterView />
        </div>
      </div>
    </div>
  </main>
</template>
