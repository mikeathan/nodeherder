<script setup lang="ts">
import { TimeValidationFunction } from '@/types/time-picker.types';
import { computed, PropType, ref } from 'vue';

const props = defineProps({
  value: {
    type: Date,
    default: '',
    required: true,
  },
  label: {
    type: String,
    default: '',
    required: false,
  },
  disabled: {
    type: Boolean,
    default: false,
    required: false,
  },
  validation: {
    type: Function as PropType<TimeValidationFunction>,
    default: () => { },
    required: false,
  },
});

const selectedTime = ref<Date>(props.value);
const errorMessage = ref<string>('');
const hasError = computed(() => {
  return errorMessage.value != '';
});

const setError = (message: string) => {
  errorMessage.value = message;
};

defineExpose({
  setError,
});
const emit = defineEmits<{
  (e: 'updated', value: any): void;
}>();

function handleEnterKey() {
  if (props.validation) {
    var res = props.validation(selectedTime.value);
    if (res) {
      setError(res);
      return;
    }
  }
  setError('');
  emit('updated', selectedTime.value);
}
</script>

<template>
  <IftaLabel>
    <DatePicker inputId="date" id="datepicker-timeonly" v-model="selectedTime" showIcon fluid iconDisplay="input"
      timeOnly @blur="handleEnterKey" :disabled="props.disabled">
      <template #inputicon="slotProps">
        <i class="pi pi-clock" @click="slotProps.clickCallback" />
        <p v-if="hasError">{{ errorMessage }}</p>
      </template>
    </DatePicker>
    <label v-if="label" for="date">{{ props.label }}</label>
  </IftaLabel>
</template>
