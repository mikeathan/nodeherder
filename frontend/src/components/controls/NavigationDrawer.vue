<script setup lang="tsx">
  import { computed, PropType, ref } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';

  const props = defineProps({
    items: {
      type: Object as PropType<MenuBarItem[]>,
      default: [],
      required: true,
    },
  });

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

  const logoItem = computed(() => {
    return props.items.find((item) => item.isLogo);
  });

  const visible = ref(true);
  const minimized = ref(false);
  function toggleMinimize() {
    minimized.value = !minimized.value;
  }
</script>

<script setup lang="tsx">
import { computed, PropType, ref } from 'vue';
import { MenuBarItem } from '@/types/controls.type';

const props = defineProps({
  items: {
    type: Object as PropType<MenuBarItem[]>,
    default: () => [],
    required: true,
  },
});

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

const minimized = ref(false);
function toggleMinimize() {
  minimized.value = !minimized.value;
}
</script>

<style scoped>
.custom-sidebar {
  background-color: #1b1b1b;
  border-right: 1px solid rgba(255, 255, 255, 0.1);
  transition: width 0.3s ease;
  overflow: hidden;
  color: white;
  height: 100vh;
  position: relative;
}

.logo-container {
  height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  position: relative;
}

.toggle-btn {
  position: absolute;
  right: 10px;
  top: 10px;
  cursor: pointer;
  background: none;
  border: none;
  color: white;
  font-size: 18px;
}

.menu-panel.minimized .p-menuitem-text {
  display: none !important; /* hide labels */
}

.menu-panel.minimized .p-menuitem-icon {
  margin: 0 auto !important; /* center icons */
  display: block !important;
}
</style>

<template>
  <Sidebar
    :visible="true"
    position="left"
    :style="{ width: minimized ? '60px' : '250px' }"
    class="custom-sidebar"
    :modal="false"
    :dismissable="false"
    :showCloseIcon="false"
  >
    <div class="logo-container">
      <component :is="logoItem?.template" />
      <button class="toggle-btn" @click="toggleMinimize">
        {{ minimized ? '▶' : '◀' }}
      </button>
    </div>
    <PanelMenu :model="menuItems" :class="{ minimized: minimized }" class="menu-panel" />
  </Sidebar>
</template>