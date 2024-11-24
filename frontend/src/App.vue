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
</style>

<template>

  <main>
    <div class="app-container">
      <Menubar :model="menuItems" class="top-navbar">
        <template #start>
          <div class="menu-links">
            <Status />
            <Notifications />
          </div>
        </template>
      </Menubar>
      <!-- <Menubar class="top-navbar">
        <template #start>
          <div class="menu-links">
            <Status></Status>
            <Notifications />

            <RouterLink to="/" class="menu-link">{{ title }}</RouterLink>
            <RouterLink to="/viewer" class="menu-link">Automations </RouterLink>
            <RouterLink to="/consoleviewer" class="menu-link">Console</RouterLink>
            <RouterLink to="/settings" class="menu-link">Settings </RouterLink>
          </div>
        </template>
</Menubar> -->

      <div class="content">
        <RouterView />
      </div>
    </div>
  </main>
  <!-- <main className="content p-0 p-sm-3">
      
      <div class="container-fluid p-0 h-100">
        <span class="me-1">
          <Status></Status>
        </span>
        <Notifications />
        <RouterLink to="/">{{ title }}</RouterLink> |
        <RouterLink to="/viewer">Automations </RouterLink> |
        <RouterLink to="/consoleviewer">Console</RouterLink> |
        <RouterLink to="/settings">Settings </RouterLink>

        <RouterView />
      </div>
    </main> -->
</template>
