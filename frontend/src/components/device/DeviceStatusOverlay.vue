<script setup lang="ts">
  import { PropType } from 'vue';
  import Icon from '@/components/controls/Icon.vue';
  import type { IconProps } from '@/types/icon.type';
import { mdiSizeL } from '@mdi/js';

  const props = defineProps({
    icon: {
      type: Object as PropType<IconProps>,
      required: true,
    },
    text: {
      type: String,
      default: '',
    },
    size: {
      type: Number,
      default: 32,
    },
    rounded: {
      type: Boolean,
      default: false,
    },
  });
</script>

<template>
  <div class="device-status-overlay">
    <Icon :icon="icon" :size="size" :circle-radius="30" :clickable="rounded" background="#363636" />
    <p v-if="text" class="status-text">{{ text }}</p>
  </div>
</template>ext="Device is disabled" 

<style scoped>
  .device-status-overlay {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    backdrop-filter: blur(2px);
    border-radius: 1rem;
    color: #777;
    gap: 0.4rem;
  }

  .status-text {
    font-size: 0.95rem;
    font-weight: 500;
  }
</style>
