<script setup lang="tsx">
  import { ComponentPublicInstance, PropType, ref } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';
  import { useWindowSize } from '@/mixins/composables/useWindowsSize';
  import { useMenuItems } from '@/mixins/composables/useMenuItems';
  import { useAuth } from '@/mixins/composables/useAuthentication';
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

  const { menuItems, logoItem } = useMenuItems(props.items);
  const { isMobile } = useWindowSize();

  const menubarRef = ref<ComponentPublicInstance | null>(null);
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
          <div class="app-version">v{{ version }}</div>
          <i v-if="isAuthenticated" class="pi pi-sign-out sign-out" @click="signOut" />
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
    color: #666;
    user-select: none;
    pointer-events: none;
    opacity: 0.7;
  }
  .sign-out {
    font-size: 1rem;
    color: #666;
    cursor: pointer;
    opacity: 0.7;
  }
  .navigation-wrapper {
    position: relative;
    user-select: none;
  }
</style>
