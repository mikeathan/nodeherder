<script setup lang="ts">
  import { computed, ref, PropType, onMounted, onUnmounted, watch } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';

  const props = defineProps({
    items: {
      type: Object as PropType<MenuBarItem[]>,
      default: () => [],
      required: true,
    },
    isMinimised: {
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
  const minimized = ref(props.isMinimised);

  const windowWidth = ref(window.innerWidth);
  const isMobile = computed(() => windowWidth.value < 768);

  const toggleMinimize = () => {
    minimized.value = !minimized.value;
    emit('widthChanged', minimized.value ? 60 : 250);
  };

  const drawerWidth = computed(() => {
    if (isMobile.value) return 0;
    return minimized.value ? 60 : 250;
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
  watch([drawerWidth, isMobile, minimized], emitDrawerWidth);
</script>
<template>
  <div v-if="!isMobile" :class="['floating-sidebar', { minimized }]">
    <div class="top-bar">
      <component :is="logoItem?.template" v-if="!minimized" />
      <Button icon="pi pi-bars" @click="toggleMinimize" rounded text />
    </div>

    <div class="menu-area">
      <!-- Minimized buttons -->
      <div v-if="minimized" class="minimized-buttons">
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
</style>
