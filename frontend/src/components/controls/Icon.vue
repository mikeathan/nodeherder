<script setup lang="ts">
  import { computed, PropType, ref, watch } from 'vue';
  import { IconProps } from '../../types/icon.type';

  const props = defineProps({
    icon: {
      type: Object as PropType<IconProps>,
      default: {} as IconProps,
      required: true,
    },
    rotationAngle: {
      type: Number,
      default: 0,
      required: false,
    },
    size: {
      type: Number,
      default: 24,
    },
    background: {
      type: String,
      default: 'transparent',
    },
    clickable: {
      type: Boolean,
      default: false,
    },
    circleRadius: {
      type: Number,
      default: 24,
    },
  });

  const emit = defineEmits<{
    (e: 'click', event: MouseEvent): void;
  }>();

  function handleClick(event: MouseEvent) {
    if (props.clickable) {
      // Stop the event from bubbling up to parent elements
      // only when the icon itself is handling the click.
      event.stopPropagation();
      emit('click', event);
    }
  }
  const wrapperStyle = computed(() => {
    if (!props.clickable) return {};
    if (props.circleRadius === 0) return {};

    const diameter = `${props.circleRadius * 2}px`;

    return {
      width: diameter,
      height: diameter,
      borderRadius: '50%',
      border: `2px solid ${props.background}`,
      backgroundColor: props.background,
      transition: 'background-color 0.4s ease, border-color 0.4s ease',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      cursor: 'pointer',
    };
  });
</script>
<style scoped>
  .icon-wrapper {
    display: inline-flex;
    justify-content: center;
    align-items: center;
  }
  .icon-wrapper.clickable {
    cursor: pointer;
    transition:
      background-color 0.2s ease,
      border-color 0.2s ease;
  }

  .icon-wrapper.clickable:hover {
    box-shadow: none;
  }

  .icon-wrapper.clickable:active {
    background-color: #4e4e4e !important;
    border-color: #4e4e4e !important;
  }
</style>
<template>
  <div class="icon-wrapper" :class="{ clickable }" @click="handleClick" :style="wrapperStyle">
    <svg :width="props.size" :height="props.size" viewBox="0 0 24 24" :aria-label="props.icon.tooltip">
      <title>{{ props.icon.tooltip }}</title>

      <g :transform="`rotate(${rotationAngle} 12 12) scale(0.8)`">
        <path :d="props.icon.name" :fill="props.icon.color || 'black'" transform="translate(2.4, 2.4)" />
      </g>
    </svg>
  </div>
</template>
