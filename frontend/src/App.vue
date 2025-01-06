<script setup lang="ts">
  import { onBeforeMount, ref, h, resolveComponent, computed } from 'vue';
  import { store } from './store/index';
  import Status from './components/controls/Status.vue';
  import Notifications from './components/hub/alerts/Notifications.vue';
  import { useRouter } from 'vue-router';
  import TimerButton from './components/controls/TimerButton.vue';
  import { Button } from 'primevue';
  import MenuBar from '@/components/controls/MenuBar.vue';
  import Logo from '@/components/controls/Logo.vue';

  onBeforeMount(() => {
    store.dispatch('ws/connect');
  });
  const router = useRouter();
  const menuItems = [
    {
      to: '/',
      label: 'dashboard',
      icon: 'pi pi-home',
      command: () => router.push('/'),
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
      label: 'Test',
      icon: 'pi pi-cog',
      template: () => h(TimerButton, { duration: 60 }),
    },
  ];
</script>
<style scoped>
  body {
    font-family: 'Roboto', sans-serif !important;
  }

  .p-component {
    font-family: 'Roboto', sans-serif !important;
  }

  .logo-container {
    position: relative;
    display: inline-flex;
    align-items: center;
  }

  .logo-text {
    font-family: 'Roboto';
    color: var(--bs-body-color);
    font-size: 24px;
    font-weight: 700;
    margin-left: 10px;
    white-space: nowrap;
  }

  .status-icon {
    position: absolute;
    top: 15px;
    right: -10px;
  }
</style>

<!-- <div class="logo-container">
  <RouterLink :to="`/`" style="text-decoration: none">
    <img src="./assets/images/nodeherder_logo.png" width="50" height="50" alt="Node-Herder"
      class="logo-image" />
    <Status class="status-icon" />
    <span class="logo-text">Node-Herder</span>
  </RouterLink>

</div> -->
<template>
  <main>
    <div class="app-container">
      <div class="col-12">
        <MenuBar :items="menuItems" />
        <Notifications />

        <div class="content">
          <RouterView />
        </div>
      </div>
    </div>
  </main>
</template>
