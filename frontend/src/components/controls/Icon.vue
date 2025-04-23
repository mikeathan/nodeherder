<script setup lang="ts">
  import { PropType, ref, watch } from 'vue';
  import { IconProps } from '../../types/icon.type';

  const props = defineProps({
    icon: {
      type: Object as PropType<IconProps>,
      default: {} as IconProps,
      required: true,
    },
    rotationAngle: {
      type: String,
      default: '0',
      required: false,
    },
    size: {
      type: String,
      default: '20px',
    },

    background: {
      type: String,
      default: '#ccc',
    },
  });
</script>
<style scoped></style>
<template>
  <svg :width="props.size" :height="props.size" viewBox="0 0 24 24" :aria-label="props.icon.tooltip">
    <title>{{ props.icon.tooltip }}</title>

    <!-- Circle background -->
    <circle cx="12" cy="12" r="12" :fill="background" />

    <!-- Icon path, centered and scaled -->
    <g :transform="`rotate(${rotationAngle} 12 12) scale(0.8)`">
      <path :d="props.icon.name" :fill="props.icon.color || 'black'" transform="translate(2.4, 2.4)" />
    </g>
  </svg>
</template>
