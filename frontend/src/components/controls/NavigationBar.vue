<script setup lang="tsx">
  import { ComponentPublicInstance, PropType, ref } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';
  import { useWindowSize } from '@/mixins/composables/useWindowsSize';
  import { useMenuItems } from '@/mixins/composables/useMenuItems';
  import { useAuth } from '@/mixins/composables/useAuthentication';
import { store } from '@/store';
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
  const { user, isAuthenticated } = useAuth();

  const toggleMobileMenu = () => {
    mobileMenuActive.value = !mobileMenuActive.value;

    const el = menubarRef.value?.$el || null;
    if (el) {
      el.classList.toggle('p-menubar-mobile-active', mobileMenuActive.value);
    }
  };
  const { menuItems, logoItem } = useMenuItems(props.items, toggleMobileMenu);
  const { isMobile } = useWindowSize();

  const menubarRef = ref<ComponentPublicInstance | null>(null);
  const mobileMenuActive = ref(false);

  const logout = () => {
    store.dispatch('auth/logoutUser');
  };
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
      <template #end v-if="isMobile && menuItems.length > 0">
        <i class="pi pi-ellipsis-v right-menu" @click="toggleMobileMenu" />
      </template>
    </Menubar>
    <div class="app-version">v{{ version }}</div>
    <div v-if="isAuthenticated" class="user-info">
      {{ user?.username }}
      <Button @click="logout">Logout</Button>
    </div>
    <div v-if="isMobile && mobileMenuActive" class="drawer-overlay" @click="toggleMobileMenu"></div>
  </div>
</template>

<style scoped>
  .drawer-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100vw;
    height: 100vh;
    z-index: 998;
  }
  .custom-menubar {
    justify-content: space-between !important;
    z-index: 999;
  }
  .custom-menubar :deep(.p-menubar-button) {
    display: none !important;
  }
  .left-menu {
    display: flex;
    align-items: center;
    padding-right: 1rem;
    cursor: pointer;
    border-radius: 4px;
  }
  .right-menu {
    display: flex;
    align-items: center;
    cursor: pointer;
    border-radius: 4px;
  }
  .app-version {
    position: absolute;
    right: 1rem;
    top: 50%;
    transform: translateY(-50%);
    font-size: 0.875rem;
    color: #666;
    user-select: none;
    pointer-events: none;
    opacity: 0.7;
  }
  .navigation-wrapper {
    position: relative;
    user-select: none;
  }
</style>
