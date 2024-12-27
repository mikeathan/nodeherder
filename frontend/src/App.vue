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
<style>
body {
  font-family: 'Roboto', sans-serif !important;
}
.p-component {
  font-family: 'Roboto', sans-serif !important;
}
</style>
<template>
  <main>
    <div class="app-container grid  align-items-center justify-content-center">
      <div class="col-12 md:col-10 lg:col-12 p-4">
        <Menubar :model="menuItems">
          <template #start>
            <div class="menu-title">
              <Status />
              <RouterLink :to="`/`">
                <img src="./assets/images/nodeherder_logo.png" width="50" height="50" />
              </RouterLink>
              Node-Herder
              <Notifications />
            </div>
          </template>
          <!-- <template #end>
          </template> -->
        </Menubar>
        <div class="content">
          <RouterView />
        </div>
      </div>
    </div>
  </main>
</template>
