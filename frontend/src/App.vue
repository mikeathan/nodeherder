<script setup lang="ts">
import { onBeforeMount, ref } from 'vue';
import { store } from './store/index';
import Status from './components/controls/Status.vue';
import Notifications from './components/hub/alerts/Notifications.vue';
import { useRouter } from 'vue-router';
import { MenuItem } from 'primevue/menuitem';

onBeforeMount(() => {
  store.dispatch('ws/connect');
});
const router = useRouter();
const menuItems: MenuItem[] = [
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

.p-menubar {
  display: flex;
  justify-content: space-between;
}

.node-herder-text {
  font-family: 'Roboto';
  font-size: 24px; /* Adjust the font size for prominence */
  font-weight: 700; /* Bold weight for more prominence */
  margin-left: 10px; /* Space between logo and text */

}

/* Optional: Add hover effect for more interactivity */
.node-herder-text:hover {
  cursor: pointer;
}
</style>
<template>
  <main>
    <div
      class="app-container grid align-items-center justify-content-center">
      <div class="col-12 md:col-10 lg:col-12 p-4">
        <Menubar :model="menuItems" breakpoint="400px">
          <template #start>
            <div class="menu-title">
              <Status />
              <RouterLink :to="`/`">
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
