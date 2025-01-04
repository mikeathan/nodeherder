<script setup lang="ts">
import { onBeforeMount, ref, h, resolveComponent, computed } from 'vue';
import { store } from './store/index';
import Status from './components/controls/Status.vue';
import Notifications from './components/hub/alerts/Notifications.vue';
import { useRouter } from 'vue-router';
import TimerButton from './components/controls/TimerButton.vue';
import { Button } from 'primevue';

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
    template: () =>
      h(TimerButton, { duration: 60 }),
  },
  // {
  //   template: () => {
  //     return {
  //       render() {
  //         const TimerButtonComponent = resolveComponent('TimerButton');
  //         return h(TimerButtonComponent, { duration: 60 });
  //       },
  //     };
  //   },
  //},
];

const isMenuVisible = ref(true);
const toggleMenu = () => {
  isMenuVisible.value = !isMenuVisible.value;
  console.log('Menu visibility:', isMenuVisible.value);
};
const standardItems = computed(() => {
  // Filter out standard items
  return menuItems.filter(item => !item.custom);
});
const customItems = computed(() => {
  // Filter out standard items
  return menuItems.filter(item => item.custom);
});
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

.p-menubar-root-list {
  display: flex;
  list-style: none;
  margin: 0;
  padding: 0;
}

.p-menuitem-link {
  display: flex;
  align-items: center;
  padding: 0.5rem 1rem;
  text-decoration: none;
  color: inherit;
  cursor: pointer;
}

.p-menuitem-link:hover {
  background-color: var(--surface-hover);
  color: var(--primary-color);
}

.p-menuitem-icon {
  margin-right: 0.5rem;
}

.custom-menubar {
  position: relative;
}

/* Hamburger Button */
.hamburger-button {
  display: none;
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: var(--text-color);
  position: absolute;
  top: 0.5rem;
  right: 1rem;
}

/* Show hamburger button on mobile */
@media (max-width: 768px) {
  .hamburger-button {
    display: block;
  }

  /* Initially hide Menubar in mobile */
  .p-menubar {
    display: none;
  }

  /* Show Menubar when it's toggled */
  .mobile-menu-visible {
    display: block;
  }

  /* Ensure mobile view stacks menu items vertically */
  .p-menubar-root-list {
    flex-direction: column;
  }

  .p-menuitem {
    width: 100%;
  }
}
/* .p-menubar {
  display: flex;
  align-items: center;
  justify-content: space-between;
}



@media (max-width: 970px) {


  .p-menubar {
    display: flex;
    align-items: center;
    flex-direction: row-reverse;
    justify-content: space-between;
  }
} 
*/
</style>

<!-- <div class="logo-container">
  <RouterLink :to="`/`" style="text-decoration: none">
    <img src="./assets/images/nodeherder_logo.png" width="50" height="50" alt="Node-Herder"
      class="logo-image" />
    <Status class="status-icon" />
    <span class="logo-text">Node-Herder</span>
  </RouterLink>

</div>
<Notifications /> -->
<template>
  <main>
    <div class="app-container">
      <div class="col-12">
        <div class="custom-menubar">
          <!-- Hamburger Icon -->
          <button class="hamburger-button" @click="toggleMenu()" aria-label="Toggle Menu">
            <i class="pi pi-bars"></i>
          </button>
          <Menubar v-if="isMenuVisible">
            <template #start>
              <ul class="p-menubar-root-list">
                <li v-for="(item, index) in menuItems" :key="index" class="p-menuitem">
                  <!-- Standard Menu Items -->
                  <template v-if="!item.custom">
                    <a class="p-menuitem-link" href="javascript:void(0)" @click="item.command && item.command()">
                      <span v-if="item.icon" :class="['p-menuitem-icon', item.icon]"></span>
                      <span class="p-menuitem-text">{{ item.label }}</span>
                    </a>
                  </template>

                  <!-- Custom Menu Items -->
                  <template v-else>
                    <component v-if="item.template" :is="item.template()" />
                  </template>
                </li>
              </ul>
            </template>
          </Menubar>
        </div>
        <div class="content">
          <RouterView />
        </div>
      </div>
    </div>
  </main>
</template>
