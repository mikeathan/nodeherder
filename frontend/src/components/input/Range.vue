<script setup lang="ts">
  import { InputNumberInputEvent, SliderSlideEndEvent } from 'primevue';
  import { ref, watch, watchEffect } from 'vue';

  const emit = defineEmits<{
    (e: 'update', id: number): void;
  }>();

  export interface Props {
    value: any | null;
    min?: number;
    max?: number;
    showInput?: boolean;
    disabled?: boolean;
  }

  const props = withDefaults(defineProps<Props>(), {
    value: 0,
    min: 0,
    max: 254,
    showInput: false,
  });

  const localValue = ref<number>();
  const localMax = ref<number>(0);
  const localMin = ref<number>(254);

  watchEffect(() => (localMin.value = props.min));
  watchEffect(() => (localMax.value = props.max));

  watch(
    () => props.value,
    () => {
      if (props.value == null || typeof props.value != 'number') {
        localValue.value = 0;
      } else {
        localValue.value = props.value;
      }
    },
    { immediate: true }
  );

  function onSlideEnds(event: SliderSlideEndEvent) {
    emit('update', event.value as number);
  }

  function inputChanged(event: InputNumberInputEvent) {
    var v = parseInt(event.value?.toString() || '0');
    emit('update', v);
  }
</script>
<style scoped></style>
<template>
  <InputNumber
    v-model.number="localValue"
    v-if="props.showInput"
    @input="inputChanged"
    :disabled="props.disabled"
    class="w-full mb-4" />
  <Slider
    v-model="localValue"
    @slideend="onSlideEnds"
    :max="localMax"
    :min="localMin"
    :disabled="props.disabled"
    class="w-full" />
</template>
