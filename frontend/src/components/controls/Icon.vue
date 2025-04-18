<script setup lang="ts">
  import { PropType } from 'vue';
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
    width: {
      type: String,
      default: '20px',
    },
    height: {
      type: String,
      default: '20px',
    },
    circleColor: {
      type: String,
      default: '#ccc',
    },
  });
</script>
<style scoped>
  /* Temporary */
  .align-middle {
    margin-right: 0.2em;
  }
</style>
<template>
  <svg
    :width="width"
    :height="height"
    viewBox="0 0 22 22"
    :style="{ fill: props.icon.color }"
    :aria-label="props.icon.tooltip"
    class="align-middle">
    <title>{{ props.icon.tooltip }}</title>
    <!-- <circle cx="11" cy="11" r="11" stroke="currentColor" fill="none" stroke-width="1" /> -->
    <g :transform="`rotate(${rotationAngle} 12 12)`">
      <path :d="props.icon.name" />
    </g>
  </svg>
</template>
