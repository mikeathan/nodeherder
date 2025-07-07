<script setup lang="ts">
  import { PropType } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';
  import { useMenuItems } from '@/mixins/composables/useMenuItems';
  import { useDrawer } from '@/mixins/composables/useDrawer';

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
    (e: 'widthChanged', width: number): void;
    (e: 'toggle'): void;
  }>();

  const { menuItems } = useMenuItems(props.items);
  const { isMobile } = useDrawer(
    () => props.isExpanded,
    (w) => emit('widthChanged', w)
  );

  const toggleDrawer = () => emit('toggle');
  const closeDrawer = () => emit('toggle');

  function handleOutsideClick() {
    console.log('outsideClose');
    if (props.isExpanded) {
      console.log('closeDrawer');
      closeDrawer();
    }
  }
</script>

<template>
  <!-- Desktop Sidebar -->
  <div v-show="!isMobile" :class="['floating-sidebar', { expanded: isExpanded }]" v-click-outside="handleOutsideClick">
    <div class="top-bar">
      <Button :icon="isExpanded ? 'pi pi-times' : 'pi pi-bars'" @mousedown="toggleDrawer" rounded text />
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
    <div v-show="isMobile && isExpanded" class="mobile-drawer" v-click-outside="handleOutsideClick">
      <div class="top-bar">
        <Button icon="pi pi-times" @mousedown="closeDrawer" rounded text />
      </div>
      <PanelMenu :model="menuItems" class="menu-panel" />
    </div>
  </transition>
</template>

<style scoped>
  .floating-sidebar {
    position: fixed;
    top: 0;
    left: 0;
    height: 100vh;
    width: 60px;
    background-color: #1b1b1b;
    color: white;
    border-right: 1px solid rgba(255, 255, 255, 0.1);
    transition: width 0.3s ease;
    display: flex;
    flex-direction: column;
    z-index: 1000;
  }

  .floating-sidebar.expanded {
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
