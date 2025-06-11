<script setup lang="tsx">
  import { ComponentPublicInstance, computed, onMounted, onUnmounted, PropType, ref } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';

  const props = defineProps({
    items: {
      type: Object as PropType<MenuBarItem[]>,
      default: [],
      required: true,
    },
  });
  const emit = defineEmits<{
    (e: 'minimised', value: boolean): void;
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

  const logoItem = computed(() => {
    return props.items.find((item) => item.isLogo);
  });
  const menubarRef = ref<ComponentPublicInstance | null>(null);
  const mobileMenuActive = ref(false);

  const toggleMobileMenu = () => {
    console.log('toggleMobileMenu');
    mobileMenuActive.value = !mobileMenuActive.value;

    const el = menubarRef.value?.$el || null;
    if (el) {
      el.classList.toggle('p-menubar-mobile-active', mobileMenuActive.value);
    }
  };

  function closeMobileMenu() {
    console.log('closeMobileMenu');
    mobileMenuActive.value = false;
    menubarRef.value?.$el?.classList.remove('p-menubar-mobile-active');
  }

  const windowWidth = ref(window.innerWidth);
  const isMobile = computed(() => windowWidth.value < 768);

  const isMinimised = ref(false);

  const onResize = () => {
    windowWidth.value = window.innerWidth;
  };

  const toggleExpanded = () => {
    console.log('toggleExpanded');
    isMinimised.value = !isMinimised.value;
    emit('minimised', isMinimised.value);
  };
  
  onMounted(() => {
    window.addEventListener('resize', onResize);
  });

  onUnmounted(() => {
    window.removeEventListener('resize', onResize);
  });
</script>



<template>
  
  <Menubar ref="menubarRef" :model="menuItems" class="custom-menubar">
    <template #start>
      <div v-if="isMobile">
        <i  :class="isMinimised ? 'pi pi-bars' : 'pi pi-times'" class="left-menu" @click="toggleExpanded" />
      </div>
      <component :is="logoItem?.template" />
    </template>
    <template #end v-if="isMobile">
       <div>
        {{ isMinimised }}
      </div>
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
  }
  .right-menu {
    display: flex;
    align-items: center;
    cursor: pointer;
  }
</style>