<script setup lang="ts">
  import { onBeforeMount, ref } from 'vue';
  import { store } from './store/index';
  import Status from './components/controls/Status.vue';
  import Notifications from './components/hub/alerts/Notifications.vue';
  import { useRouter } from 'vue-router';

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
  ];
</script>
<style scoped>
  body {
    font-family: 'Roboto', sans-serif !important;
  }

  .p-component {
    font-family: 'Roboto', sans-serif !important;
  }

  .menu-title {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .node-herder-text {
    font-family: 'Roboto';
    color: var(--bs-body-color);
    font-size: 24px;
    font-weight: 700;
    margin-left: 10px;
    white-space: nowrap;
  }

  @media (max-width: 970px) {
    /* .node-herder-text {
    display: none;
  } */

    .menu-title {
      flex: 1;
    }

    .p-menubar {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
  }
</style>
<template>
  <main>
    <div class="app-container">
      <div class="col-12">
        <Menubar :model="menuItems">
          <template #start>
            <div class="menu-title">
              <Status />
              <RouterLink
                :to="`/`"
                style="text-decoration: none">
                <img
                  src="./assets/images/nodeherder_logo.png"
                  width="50"
                  height="50" />
                <span class="node-herder-text"
                  >Node-Herder</span
                >
              </RouterLink>

              <Notifications />
            </div>
          </template>
        </Menubar>
        <div class="content">
          <RouterView />
        </div>
      </div>
    </div>
  </main>
</template>
