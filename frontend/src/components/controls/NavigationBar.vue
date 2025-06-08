<script setup lang="tsx">
  import { computed, PropType } from 'vue';
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
