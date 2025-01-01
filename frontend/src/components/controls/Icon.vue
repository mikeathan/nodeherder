<script setup lang="ts">
  import { PropType } from 'vue';
  import { IconProps } from '../../types/icon.type';

  const props = defineProps({
    icon: {
      type: Object as PropType<IconProps>,
      default: {} as IconProps,
      required: true,
    },
  });
</script>
<style scoped>
  .icon {
    width: 22px;
    height: 22px;
    vertical-align: middle;
    margin-right: 0.2em;
  }
</style>
<template>
  <svg
    viewBox="0 0 24 24"
    class="icon"
    :style="{ fill: props.icon.color }"
    :aria-label="props.icon.tooltip">
    <title>{{ props.icon.tooltip }}</title>
    <path :d="props.icon.name" />
  </svg>
</template>
