<script setup lang="ts">
import { onBeforeMount, ref } from 'vue';
import { store } from './store/index';
import Status from './components/controls/Status.vue';
import Notifications from './components/hub/alerts/Notifications.vue';
import { useRouter } from 'vue-router';
import TimerButton from './components/controls/TimerButton.vue';

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

.p-menubar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}



@media (max-width: 970px) {
  /* .node-herder-text {
    display: none;
  } */

  .p-menubar {
    display: flex;
    align-items: center;
    flex-direction: row-reverse;
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
            <div class="logo-container">
              <RouterLink :to="`/`" style="text-decoration: none">
                <img src="./assets/images/nodeherder_logo.png" width="50" height="50" alt="Node-Herder"
                  class="logo-image" />
                <Status class="status-icon" />
                <span class="logo-text">Node-Herder</span>
              </RouterLink>

            </div>
            <Notifications />
            <!-- <div class="menu-end-container">
              <TimerButton :duration="60" />
            </div> -->
            <!-- <div class="menu-end-container"> 
            <TimerButton
              :duration="60"
              :start-event="
                () => {
                  console.log('start event');
                }
              "
              :stop-event="
                () => {
                  console.log('stop event');
                }
              " />
              </div> -->
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
