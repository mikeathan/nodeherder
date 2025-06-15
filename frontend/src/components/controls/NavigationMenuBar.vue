<script setup lang="tsx">
  import {  PropType } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';
  import { useMenuItems } from '@/mixins/composables/useMenuItems';

  const props = defineProps({
    items: {
      type: Object as PropType<MenuBarItem[]>,
      default: [],
      required: true,
    },
  });

  const { menuItems, logoItem } = useMenuItems(props.items);
</script>

<style scoped>
  .custom-menubar {
    justify-content: space-between !important;
  }
</style>

<template>
  <Menubar :model="menuItems" class="custom-menubar">
    <template #start>
      <component :is="logoItem?.template" />
    </template>
  </Menubar>
</template>
