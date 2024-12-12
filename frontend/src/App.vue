<script setup>
import {
  onBeforeMount,
  onMounted,
  onUnmounted,
  ref,
} from 'vue';
import { store } from './store/index';
import Status from './components/controls/Status.vue';
import Notifications from './components/hub/alerts/Notifications.vue';
import InputText from 'primevue/inputtext';
import Button from 'primevue/button';
const title = ref('Node-Herder');
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
<!-- 
<style>
app-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
}

.top-navbar {
  position: sticky;
  top: 0;
  z-index: 1000;
  background-color: var(--primary-color, #007ad9);
  color: white;
  box-shadow: 0px 2px 4px rgba(0, 0, 0, 0.1);
  padding: 0 1rem;
}

.menu-links {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.menu-link {
  text-decoration: none;
  color: white;
  font-weight: 500;
}

.menu-link:hover {
  text-decoration: underline;
}

.content {
  flex: 1;
  padding: 1rem;
  overflow: auto;
}
</style> -->

<template>

  <main class="">
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
