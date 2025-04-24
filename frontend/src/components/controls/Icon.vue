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
    clickable: {
      type: Boolean,
      default: false,
    },
  });

  const emit = defineEmits<{
    (e: 'click'): void;
  }>();

  function handleClick() {
    if (props.clickable) {
      emit('click');
    }
  }
</script>
<style scoped>
  svg.clickable {
    cursor: pointer;
  }

  svg.clickable:hover circle {
    opacity: 0.6;
    transition: opacity 0.2s ease;
  }
</style>
<template>
  <svg
    :width="props.size"
    :height="props.size"
    viewBox="0 0 24 24"
    :aria-label="props.icon.tooltip"
    @click.stop="handleClick"
    :class="{ clickable }">
    <title>{{ props.icon.tooltip }}</title>

    <circle cx="12" cy="12" r="12" :fill="background" v-if="clickable" />
    <!-- Icon path, centered and scaled -->
    <g :transform="`rotate(${rotationAngle} 12 12) scale(0.8)`">
      <path :d="props.icon.name" :fill="props.icon.color || 'black'" transform="translate(2.4, 2.4)" />
    </g>
  </svg>
</template>
