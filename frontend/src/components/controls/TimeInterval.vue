<script setup lang="ts">
  import { PropType, ref, watchEffect } from 'vue';
  import InputBox from '../input/InputBox.vue';
import { TimeInterval } from '../../types/types.type';

  const props = defineProps({
    id: { type: String, required: true },
    value: {
      type: Object as PropType<TimeInterval>,
      required: true,
    },
  });
  const emit = defineEmits<{
    (e: 'update', value: TimeInterval): void;
  }>();

  const timeInterval = ref<TimeInterval>(props.value);
    watchEffect(() => (timeInterval.value = props.value));

  function inputTimeIntervalLostFocus(propValue: any) {
    timeInterval.value.value = propValue;
    emit('update', timeInterval.value);
  }
</script>

<template>
  <InputBox
    :label="timeInterval.unit"
    :value="timeInterval.value"
    :is-numeric="true"
    @lost-focus="(f) => inputTimeIntervalLostFocus(f)" />
</template>
