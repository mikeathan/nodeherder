<script setup lang="tsx">
  import { ComponentPublicInstance, PropType, ref } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';
  import { useWindowSize } from '@/mixins/composables/useWindowsSize';
  import { useMenuItems } from '@/mixins/composables/useMenuItems';
  import { useAuth } from '@/mixins/composables/useAuthentication';
  import { useTheme } from '@/services/theme.service';
  const version = __APP_VERSION__;

  const props = defineProps({
    items: {
      type: Object as PropType<MenuBarItem[]>,
      default: [],
      required: true,
    },
  });

  const emit = defineEmits<{
    (e: 'click', value: boolean): void;
  }>();

  const onDrawerToggle = () => emit('click', true);
  const { user, isAuthenticated, signOut } = useAuth();
  const { isDarkMode, toggleTheme } = useTheme();

  const { menuItems, logoItem } = useMenuItems(props.items);
  const { isMobile } = useWindowSize();

</script>

<template>
  <div class="navigation-wrapper">
    <Menubar ref="menubarRef" :model="menuItems" class="custom-menubar">
      <template #start>
        <div v-if="isMobile">
          <i class="pi pi-bars left-menu" @click="onDrawerToggle" />
        </div>
        <component :is="logoItem?.template" />
      </template>
      <template #end>
        <div class="app-info">
          <i
            :class="['pi', isDarkMode ? 'pi-moon' : 'pi-sun', 'action-icon']"
            @click="toggleTheme"
            v-tooltip.bottom="isDarkMode ? 'Switch to Light Mode' : 'Switch to Dark Mode'" />
          <div class="app-version">v{{ version }}</div>
          <i v-if="isAuthenticated" class="pi pi-sign-out action-icon" @click="signOut" v-tooltip.bottom="'Sign Out'" />
        </div>
      </template>
    </Menubar>
  </div>
</template>

<style scoped>
  .custom-menubar {
    justify-content: space-between !important;
    z-index: 999;
  }
  .custom-menubar :deep(.p-menubar-button) {
    display: none !important;
  }

  .app-info {
    position: absolute;
    right: 1rem;
    top: 50%;
    transform: translateY(-50%);
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }
  .app-version {
    font-size: 0.875rem;
    color: var(--text-color-secondary);
    user-select: none;
    pointer-events: none;
    opacity: 0.7;
  }
  .action-icon {
    font-size: 1rem;
    color: var(--text-color-secondary);
    cursor: pointer;
    opacity: 0.7;
    transition: opacity 0.2s;
  }
  .action-icon:hover {
    opacity: 1;
  }
  .navigation-wrapper {
    position: relative;
    user-select: none;
  }
</style>
