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
<style scoped>
.error-icon {
  color: #F44336 !important;
}
</style>
<template>
  <IftaLabel>
    <div class="date-picker-container" :class="{ 'has-error': hasError }">
      <DatePicker inputId="date" id="datepicker-timeonly" v-model="selectedTime" showIcon fluid iconDisplay="input"
        timeOnly @blur="handleEnterKey"
        @keyup.enter="handleEnterKey" 
         :disabled="props.disabled" :class="{ 'error-overlay': hasError }">
        <template #inputicon="slotProps">
          <i class="pi pi-clock" @click="slotProps.clickCallback" :class="{ 'error-icon': hasError }" />
        </template>
      </DatePicker>
    </div>
    <Message v-if="hasError" severity="error" size="small">{{ errorMessage }}</Message>
    <label v-if="label" for="date"> {{ props.label }}</label>
  </IftaLabel>
</template>
