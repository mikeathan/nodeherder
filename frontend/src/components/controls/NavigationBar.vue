<script setup lang="tsx">
  import { ComponentPublicInstance, onUnmounted, PropType, ref } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';
  import { useWindowSize } from '@/mixins/composables/useWindowsSize';
  import { useMenuItems } from '@/mixins/composables/useMenuItems';

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

  const { menuItems, logoItem } = useMenuItems(props.items);
  const { isMobile } = useWindowSize();

  const menubarRef = ref<ComponentPublicInstance | null>(null);
  const mobileMenuActive = ref(false);

  const onDrawerToggle = () => emit('click', true);
  const toggleMobileMenu = () => {
    mobileMenuActive.value = !mobileMenuActive.value;

    const el = menubarRef.value?.$el || null;
    if (el) {
      el.classList.toggle('p-menubar-mobile-active', mobileMenuActive.value);
    }
  };
</script>

<template>
  <Menubar ref="menubarRef" :model="menuItems" class="custom-menubar">
    <template #start>
      <div v-if="isMobile">
        <i class="pi pi-bars left-menu" @click="onDrawerToggle" />
      </div>
      <component :is="logoItem?.template" />
    </template>
    <template #end v-if="isMobile">
      <i class="pi pi-ellipsis-v right-menu" @click="toggleMobileMenu" />
    </template>
  </Menubar>
</template>

<style scoped>
  .custom-menubar {
    justify-content: space-between !important;
  }
  .custom-menubar :deep(.p-menubar-button) {
    display: none !important;
  }
  .left-menu {
    display: flex;
    align-items: center;
    padding-right: 1rem;
    cursor: pointer;
  }
  .right-menu {
    display: flex;
    align-items: center;
    cursor: pointer;
  }
</style>
