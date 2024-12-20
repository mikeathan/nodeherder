<script setup>
import {
  onBeforeMount,
} from 'vue';
import { store } from './store/index';
import Status from './components/controls/Status.vue';
import Notifications from './components/hub/alerts/Notifications.vue';
import { useRouter } from 'vue-router';

onBeforeMount(() => {
  store.dispatch('ws/connect');
});
const router = useRouter();
const menuItems = [
  { to: "/", label: "dashboard", icon: 'pi pi-home', command: () => router.push('/') },
  { to: "/viewer", label: "automations", icon: 'pi pi-objects-column', command: () => router.push('/viewer') },
  { to: "/consoleviewer", label: "console", icon: 'pi pi-code', command: () => router.push('/consoleviewer') },
  { to: "/settings", label: "settings", icon: 'pi pi-cog', command: () => router.push('/settings') },
];


</script>

<template>
  <main>
    <div class="app-container grid  align-items-center justify-content-center">
      <div class="col-12 md:col-10 lg:col-12 p-4">
        <Menubar :model="menuItems" class="top-navbar">
          <template #start>
            <div class="menu-links">
              <Status />
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
