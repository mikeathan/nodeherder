<script setup lang="ts">
  import { PropType } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';
  import { useMenuItems } from '@/mixins/composables/useMenuItems';
  import { useWindowSize } from '@/mixins/composables/useWindowsSize';

  const props = defineProps({
    items: {
      type: Object as PropType<MenuBarItem[]>,
      default: () => [],
      required: true,
    },
    isExpanded: {
      type: Boolean,
      default: false,
    },
  });

  const emit = defineEmits<{
    (e: 'toggle'): void;
  }>();

  const toggleDrawer = () => emit('toggle');
  const closeDrawer = () => emit('toggle');

  const onMenuItemClick = () => {
    if (props.isExpanded) {
      closeDrawer();
    }
  };
  const { menuItems } = useMenuItems(props.items, onMenuItemClick);
  const { isMobile } = useWindowSize();
</script>

<template>
  <!-- Desktop Sidebar -->
  <div v-if="!isMobile" :class="['sidebar', { expanded: isExpanded }]">
    <div class="top-bar">
      <Button :icon="isExpanded ? 'pi pi-times' : 'pi pi-bars'" @click="toggleDrawer" rounded text />
    </div>
    <div class="menu-area">
      <PanelMenu v-if="isExpanded" :model="menuItems" class="menu-panel" />
      <div v-else class="minimized-buttons">
        <Button
          v-for="item in menuItems"
          :key="item.label"
          :icon="item.icon"
          @click="item.command"
          :disabled="item.disabled"
          v-tooltip.right="item.label"
          text
          rounded />
      </div>
    </div>
  </div>

  <!-- Mobile Drawer -->
  <transition name="slide-left">
    <div v-if="isMobile && isExpanded" class="mobile-drawer">
      <div class="top-bar">
        <Button icon="pi pi-times" @click="closeDrawer" rounded text />
      </div>
      <PanelMenu :model="menuItems" class="menu-panel" />
    </div>
  </transition>

  <!-- Overlay to catch outside clicks -->
  <div v-if="isExpanded" class="drawer-overlay" @click="closeDrawer"></div>
</template>

<style scoped>
  .drawer-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    z-index: 999;
    background: rgba(0, 0, 0, 0.2);
  }
  .sidebar {
    width: 60px;
    background-color: #1b1b1b;
    color: white;
    border-right: 1px solid rgba(255, 255, 255, 0.1);
    transition: width 0.3s ease;
    display: flex;
    flex-direction: column;
    z-index: 1000;
  }

  .sidebar.expanded {
    width: 250px;
  }

  .top-bar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.75rem;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .menu-area {
    flex-grow: 1;
    overflow-y: auto;
    padding: 0.5rem;
  }

  .minimized-buttons {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.75rem;
    margin-top: 1rem;
  }

  .mobile-drawer {
    position: fixed;
    top: 0;
    left: 0;
    width: 250px;
    height: 100vh;
    background-color: #1b1b1b;
    color: white;
    border-right: 1px solid rgba(255, 255, 255, 0.1);
    display: flex;
    flex-direction: column;
    z-index: 1100;
    box-shadow: 2px 0 5px rgba(0, 0, 0, 0.5);
    padding: 0.5rem;
  }

  /* Slide animation */
  .slide-left-enter-active,
  .slide-left-leave-active {
    transition: transform 0.3s ease;
  }

  .slide-left-enter-from,
  .slide-left-leave-to {
    transform: translateX(-100%);
  }

  .slide-left-enter-to,
  .slide-left-leave-from {
    transform: translateX(0);
  }
</style>
