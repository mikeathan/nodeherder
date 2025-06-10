<script setup lang="ts">
  import { computed, ref, PropType, onMounted, onUnmounted, watch } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';
  import { watchEffect } from 'vue';

  const props = defineProps({
    items: {
      type: Object as PropType<MenuBarItem[]>,
      default: () => [],
      required: true,
    },
    minimised: {
      type: Boolean,
      required: false,
      default: false,
    },
  });
  const emit = defineEmits<{
    (e: 'widthChanged', width: number): void;
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

  const logoItem = computed(() => props.items.find((item) => item.isLogo));
  const isMinimised = ref(props.minimised);

  const windowWidth = ref(window.innerWidth);
  const isMobile = computed(() => windowWidth.value < 768);

  const toggleMinimize = () => {
    isMinimised.value = !isMinimised.value;
    emit('widthChanged', isMinimised.value ? 60 : 250);
  };

  const drawerWidth = computed(() => {
    if (isMobile.value) return 0;
    return isMinimised.value ? 60 : 250;
  });

  const onResize = () => {
    windowWidth.value = window.innerWidth;
  };

  onMounted(() => {
    window.addEventListener('resize', onResize);
    emitDrawerWidth();
  });

  onUnmounted(() => {
    window.removeEventListener('resize', onResize);
  });

  const emitDrawerWidth = () => {
    emit('widthChanged', drawerWidth.value);
  };
  watch([drawerWidth, isMobile, isMinimised], emitDrawerWidth);
  watchEffect(() => (isMinimised.value = props.minimised));
</script>
<template>
{{ isMinimised }}
  <div v-if="!isMobile" :class="['floating-sidebar', { isMinimised }]">
    <div class="top-bar">
      <Button :icon="isMinimised ? 'pi pi-bars' : 'pi pi-times'" @click="toggleMinimize" rounded text />
    </div>

    <div class="menu-area">
      <!-- Minimized buttons -->
      <div v-if="isMinimised" class="minimized-buttons">
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
      <!-- Full menu -->
      <PanelMenu v-else :model="menuItems" class="menu-panel" />
    </div>
  </div>
  <!-- Mobile drawer sliding in from right -->
  <transition name="slide-left">
    <div v-if="isMobile && !isMinimised" class="mobile-drawer">
      <div class="top-bar">
        <Button icon="pi pi-times" @click="toggleMinimize" rounded text />
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
    background-color: #1b1b1b;
    color: white;
    border-right: 1px solid rgba(255, 255, 255, 0.1);
    width: 250px;
    transition: width 0.3s ease;
    display: flex;
    flex-direction: column;
    z-index: 1000;
  }

  .floating-sidebar.minimized {
    width: 60px;
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
  left: 0; /* Position on left */
  width: 250px;
  height: 100vh;
  background-color: #1b1b1b;
  color: white;
  border-right: 1px solid rgba(255, 255, 255, 0.1); /* border on right side */
  display: flex;
  flex-direction: column;
  z-index: 1100;
  box-shadow: 2px 0 5px rgba(0, 0, 0, 0.5); /* shadow on right */
}

/* Animation for sliding in from left */
.slide-left-enter-active,
.slide-left-leave-active {
  transition: transform 0.3s ease;
}
.slide-left-enter-from,
.slide-left-leave-to {
  transform: translateX(-100%); /* offscreen left */
}
.slide-left-enter-to,
.slide-left-leave-from {
  transform: translateX(0); /* fully visible */
}
</style>
