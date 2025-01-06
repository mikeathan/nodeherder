<script setup lang="tsx">
  import { onBeforeMount, ref, h, resolveComponent, computed, PropType } from 'vue';
  import { store } from '../../store/index';
  import Status from './components/controls/Status.vue';
  import Notifications from './components/hub/alerts/Notifications.vue';
  import { useRouter } from 'vue-router';
  import TimerButton from './components/controls/TimerButton.vue';
  import { Button } from 'primevue';
  import { onMounted } from 'vue';
  import Logo from '@/components/controls/Logo.vue';

  type MenuItem = {
    to?: string;
    label: string;
    icon: string;
    command?: () => void;
    custom?: boolean;
    template?: () => void;
  };

  const props = defineProps({
    items: {
      type: Object as PropType<MenuItem[]>,
      default: [],
      required: true,
    },
  });

  const isMenuVisible = ref(false);
  const isMobileView = ref(false);

  const toggleMenu = () => {
    isMenuVisible.value = !isMenuVisible.value;
    console.log('Menu visibility:', isMenuVisible.value);
  };
  function checkMobileView() {
    isMobileView.value = window.innerWidth <= 768;
  }

  const renderMenuItem = (item: MenuItem) => {
    if (!item.custom) {
      return (
        <li class="menuitem">
          <a class="menuitem-link" href="javascript:void(0)" onClick={item.command}>
            {item.icon ? <span class={`menuitem-icon ${item.icon}`}></span> : null}
            <span class="menuitem-text">{item.label}</span>
          </a>
        </li>
      );
    } else if (item.template) {
      return (
        <li class="menuitem">
          <div>{item.template()}</div>
        </li>
      );
    }
    return null;
  };

  onMounted(() => {
    checkMobileView();
    window.addEventListener('resize', checkMobileView);
  });

  onBeforeMount(() => {
    window.removeEventListener('resize', checkMobileView);
  });
</script>
<style scoped>
  .menubar {
    display: flex;
    align-items: center;
    width: 100%;
    background: var(--p-menubar-background);
    color: var(--primary-text-color, white);
    border: 1px solid var(--p-menubar-border-color);
  }

  .menubar-root-list {
    display: flex;
    list-style: none;
    justify-content: flex-end;
    margin: 0;
    padding: 0;
    background: var(--p-menubar-background);
    color: var(--primary-text-color, white);
    margin-left: auto;
  }

  .menuitem-link {
    display: flex;
    align-items: center;
    margin: 0px 15px 0px 0px;
    text-decoration: none;
    color: inherit;
    cursor: pointer;
  }

  .menuitem-link:hover {
    background-color: var(--surface-hover);
    color: var(--primary-color);
  }

  .menuitem-icon {
    margin-right: 0.5rem;
  }

  @media (max-width: 768px) {
    .menubar-root-list {
      flex-direction: column;
    }

    .menuitem {
      width: 100%;
    }
  }

  .hamburger {
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--p-menubar-background);
    color: var(--primary-text-color, white);
    padding: 0.5rem;
    cursor: pointer;
  }

  .hamburger i {
    font-size: 1.5rem;
  }

  .mobile-header {
    display: flex;
    justify-content: flex-end;
    background: var(--p-menubar-background);
    color: var(--primary-text-color, white);
    border: 1px solid var(--p-menubar-border-color);
    border-radius: var(--p-menubar-border-radius);
    padding: 0.5rem;
  }
  .menubar-start {
    display: flex;
    align-items: center;
    width: 100%;
    background: var(--p-menubar-background);
    color: var(--primary-text-color, white);
    border: 1px solid var(--p-menubar-border-color);
    border-radius: var(--p-menubar-border-radius);
  }

  .menubar-logo {
    margin-right: 5rem;
  }
</style>

<template>
  <div v-if="!isMobileView">
    <div class="menubar-start">
      <div class="menubar-logo">
        <Logo />
      </div>
      <ul class="menubar-root-list">
        <li v-for="(item, index) in props.items" :key="index">
          <component :is="renderMenuItem(item)" />
        </li>
      </ul>
    </div>
  </div>

  <!-- Mobile View -->
  <div v-else>
    <div class="mobile-header">
      <div class="hamburger" @click="toggleMenu">
        <i class="pi pi-bars"></i>
      </div>
    </div>
    <transition name="menu-slide">
      <ul v-if="isMenuVisible" class="menubar-root-list">
        <li v-for="(item, index) in props.items" :key="index" class="menuitem">
          <component :is="renderMenuItem(item)" />
        </li>
      </ul>
    </transition>
  </div>
</template>
