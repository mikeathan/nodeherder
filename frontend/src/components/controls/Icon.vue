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
  });
</script>
<style scoped>
  .icon {
    width: 20px;
    height: 20px;
    vertical-align: middle;
    margin-right: 0.2em;
  }
</style>
<template>
  <svg
    viewBox="0 0 22 22"
    class="icon"
    :style="{ fill: props.icon.color }"
    :aria-label="props.icon.tooltip">
    <title>{{ props.icon.tooltip }}</title>
    <g :transform="`rotate(${rotationAngle} 12 12)`">
      <path :d="props.icon.name" />
    </g>
  </svg>
</template>
