<script setup lang="ts">
  import { computed, PropType } from 'vue';
  import InputBox from '../input/InputBox.vue';
  import { TimeInterval } from '../../types/types.type';

  const props = defineProps({
    value: {
      type: Object as PropType<TimeInterval>,
      required: true,
    },
  });
  const emit = defineEmits<{
    (e: 'update', value: TimeInterval): void;
  }>();

  const timeInterval = computed(() => props.value);

  function inputTimeIntervalLostFocus(propValue: any) {
    if (!timeInterval.value || timeInterval.value.value === propValue) return;

    timeInterval.value.value = propValue;
    emit('update', { ...timeInterval.value });
  }
</script>

<template>
  <InputBox
    :label="timeInterval.unit"
    :value="timeInterval.value"
    :is-numeric="true"
    @lost-focus="(f) => inputTimeIntervalLostFocus(f)" />
</template>
