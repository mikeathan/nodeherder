<script setup lang="ts">
import { computed, PropType, ref, onMounted, onUnmounted, watch } from 'vue';
import { MenuBarItem } from '@/types/controls.type';

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

const menuItems = computed(() =>
  props.items
    .filter((item) => !item.custom && !item.isLogo)
    .map((item) => ({
      label: item.label,
      icon: item.icon,
      disabled: item.disabled,
      command: item.command,
      items: item.children?.map((child) => ({
        label: child.label,
        icon: child.icon,
        command: child.command,
      })),
    }))
);

const windowWidth = ref(window.innerWidth);
const isMobile = computed(() => windowWidth.value < 768);
const drawerWidth = computed(() => (isMobile.value ? 0 : props.isExpanded ? 250 : 60));

const toggleDrawer = () => {
  emit('toggle');
};

const closeDrawer = () => {
  emit('toggle'); // For mobile, we still want to toggle (close)
};

const emitDrawerWidth = () => {
  emit('widthChanged', drawerWidth.value);
};

const onResize = () => {
  windowWidth.value = window.innerWidth;
  emitDrawerWidth();
};

onMounted(() => {
  window.addEventListener('resize', onResize);
  emitDrawerWidth();
});

onUnmounted(() => {
  window.removeEventListener('resize', onResize);
});

watch([() => props.isExpanded, isMobile], emitDrawerWidth);
</script>

<template>
  <!-- Desktop Sidebar -->
  <div v-if="!isMobile" :class="['floating-sidebar', { expanded: isExpanded }]">
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
          rounded
        />
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