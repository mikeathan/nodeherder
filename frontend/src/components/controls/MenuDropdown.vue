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
    text: {
      type: String,
      default: '',
      required: false,
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
    <div class="button-content" :style="{ backgroundColor }" @click="toggleMenu">
      <Icon :icon="icon" :size="size" background="transparent" />
      <span class="text-md font-medium whitespace-nowrap"> {{ text }} </span>
    </div>
    <Menu ref="menu" :model="children" popup appendTo="body" />
  </div>
</template>
<style scoped>
  .button-content {
    display: flex;
    gap: 3px;
    padding: 4px 12px;
    border-radius: 10px;
    cursor: pointer;
    transition: opacity 0.3s ease;
    align-items: center;
  }

  .button-content:hover {
    opacity: 0.8; 
  }

</style>
