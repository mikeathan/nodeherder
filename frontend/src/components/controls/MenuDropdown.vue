<script setup lang="ts">
  import { ref, PropType } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';
  import Icon from '../controls/Icon.vue';
  import Menu from 'primevue/menu';
  import { IconProps } from '@/types/icon.type';

  const props = defineProps({
    icon: {
      type: Object as PropType<IconProps>,
      default: [],
      required: true,
    },
    backgroundColor: {
      type: String,
      default: 'var(--surface-card)',
    },
    size: {
      type: Number,
      default: 38,
    },

    children: {
      type: Object as PropType<MenuBarItem[]>,
      default: [],
      required: true,
    },
  });

  const emit = defineEmits(['close']);

  function close() {
    emit('close', false);
  }

  const menu = ref<InstanceType<typeof Menu> | null>(null);
  const toggleMenu = (event: Event) => {
    menu.value?.toggle(event);
  };
</script>

<template>
  <div class="relative inline-block">
    <div
      class="flex items-center gap-2 px-3 py-2 rounded cursor-pointer hover:opacity-80 transition-opacity"
      :style="{ backgroundColor }"
      @click="toggleMenu">
      <Icon :icon="icon" :size="size" :background="'transparent'" />
      <span class="text-sm font-medium whitespace-nowrap"> Menu </span>
    </div>
    <Menu ref="menu" :model="children" popup />
  </div>
</template>
<style scoped></style>
