<script setup lang="ts">
  import { PropType } from 'vue';
  import Icon from '@/components/controls/Icon.vue';
  import type { IconProps } from '@/types/icon.type';

  const props = defineProps({
    icon: {
      type: Object as PropType<IconProps>,
      required: true,
    },
    text: {
      type: String,
      default: '',
    },
  });
</script>

<template>
  <div class="device-status-overlay">
    <Icon :icon="icon" :size="32" :circle-radius="30" :clickable="true" background="#363636" />
    <p v-if="text" class="status-text">{{ text }}</p>
  </div>
</template>

<style scoped>
  .device-status-overlay {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    min-height: 120px;
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
