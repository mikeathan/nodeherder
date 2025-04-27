<script setup lang="ts">
  import { ref, computed, watch } from 'vue';
  import { prop } from 'vue-class-component';
  const props = defineProps({
    value: {
      type: Number,
      default: 50,
    },
    min: {
      type: Number,
      default: 0,
    },
    max: {
      type: Number,
      default: 100,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  });
  const emit = defineEmits<{
    (e: 'update', value: number): void;
  }>();

  const value = ref(props.value);
  const lastKnownValue = ref(props.value);

  const trackFill = computed(() => {
    const percent = ((value.value - props.min) / (props.max - props.min)) * 100;
    const color = props.disabled ? '#ccc' : '#ffc107';
    const bgColor = props.disabled ? '#eee' : '#fff4cc';
    return {
      background: `linear-gradient(to right, ${color} ${percent}%, ${bgColor} ${percent}%)`,
    };
  });

  watch(
    () => props.value,
    (newVal) => {
      if (!props.disabled) {
        value.value = newVal;
        lastKnownValue.value = newVal;
      }
    }
  );

  watch(
    () => props.disabled,
    (isDisabled) => {
      if (isDisabled) {
        lastKnownValue.value = value.value;
        value.value = props.min;
      } else {
        value.value = lastKnownValue.value;
      }
    }
  );
  function updateValue(event: any): void {
    const newValue = parseInt(event.target.value);
    emit('update', newValue);
  }
</script>

<template>
  <div class="slider-wrapper">
    <div class="track-background" :style="trackFill"></div>
    <input
      type="range"
      :min="props.min"
      :max="props.max"
      v-model="value"
      class="slider"
      @change.stop="updateValue"
      @click.stop
      @mousedown.stop
      @pointerdown.stop />
  </div>
</template>

<style scoped>
  .slider-wrapper {
    position: relative;
    height: 46px;
    overflow: hidden;
    opacity: 1 !important;
    transition: none !important;
    
    weneed som pading
  }

  .track-background {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    border-radius: 12px;
    pointer-events: none;
   
  }

  /* input layer */
  .slider {
    appearance: none;
    width: 100%;
    height: 100%;
    background: transparent;
    position: relative;
    z-index: 1;
    cursor: pointer;
    outline: none;
  }

  /* Webkit thumb */
  /* .slider::-webkit-slider-thumb {
  appearance: none;
  width: 32px;
  height: 6px;
  background: white;
  border-radius: 3px;
  margin-top: calc((42px - 6px) / 2);
  cursor: pointer;
  position: relative;
  z-index: 2;
  box-shadow: 0 0 1px rgba(0, 0, 0, 0.2);
} */

  /* Firefox track & thumb */
  .slider::-moz-range-track {
    height: 100%;
    background: transparent;
    border: none;
  }

  .slider::-moz-range-thumb {
    width: 6px;
    height: 30px;
    background: white;
    border: none;
    border-radius: 3px;
    cursor: pointer;
    transform: translateX(-10px);
    /* Shift left by 2 pixels */
  }
  .slider:disabled {
    opacity: 0.6;
  }
</style>
