<script setup lang="tsx">
  import { computed, onMounted, onUnmounted, PropType, ref } from 'vue';
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
  const menubarRef = ref(null);
  const isMobile = computed(() => window.innerWidth < 768);
  const windowWidth = ref(window.innerWidth);
  const minimized = ref(false);

  const onResize = () => {
    windowWidth.value = window.innerWidth;
  };

  const toggleMinimize = () => {
    minimized.value = !minimized.value;
  };
  onMounted(() => {
    window.addEventListener('resize', onResize);
  });

  onUnmounted(() => {
    window.removeEventListener('resize', onResize);
  });
</script>

<style scoped>
  .custom-menubar {
    justify-content: space-between !important;
  }
  /* .custom-menubar :deep(.p-menubar-button) {
    display: none !important;
  } */
  .custom-menubar :deep(.p-menubar-button *) {
    display: none !important;
  }

  /* Add only our icon */
  .custom-menubar :deep(.p-menubar-button) {
    font-family: 'primeicons' !important;
    font-size: 1rem !important;
  }

  .custom-menubar :deep(.p-menubar-button::before) {
    content: '\e95a' !important; /* pi-ellipsis-v */
  }
</style>

<template>
  <Menubar ref="menubarRef" :model="menuItems" class="custom-menubar">
    <template #start>
      <div v-if="isMobile" :class="['floating-sidebar', { minimized }]">
        <Button :icon="minimized ? 'pi pi-times' : 'pi pi-bars'" @click="toggleMinimize" rounded text />
      </div>
      <component :is="logoItem?.template" />
    </template>
    <!-- <template #end>
      <div class="custom-mobile-menu" v-if="isMobile">
        <Button icon="pi pi-ellipsis-v" text rounded aria-label="Menu" />
      </div>
    </template> -->
  </Menubar>

  <i class="pi pi-ellipsis-v" style="font-size: 2rem; color: red;"></i>
  <!-- 
<  Button type="button" icon="pi pi-ellipsis-v" @click="toggle" aria-haspopup="true" aria-controls="overlay_menu" />
<Menu ref="menu" id="overlay_menu" :model="items" :popup="true" /> -->
</template>
