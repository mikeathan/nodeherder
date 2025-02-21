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
      default: () => {},
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
        setError('Invalid time');
        return;
      }
    }
    setError('');
    emit('updated', selectedTime.value);
  }
</script>
<style scoped>
  .date-picker-container {
    position: relative; /* Needed for positioning the *input* within the container */
    display: inline-block; /* Or block, depending on your layout */
  }

  /* Style for error border */
  .has-error .p-datepicker {
    border-color: red;
  }

  .has-error .p-datepicker:focus {
    box-shadow: 0 0 0 0.2rem rgba(255, 0, 0, 0.25);
  }

  .error-overlay {
    position: absolute; /* Position relative to the viewport */
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(255, 0, 0, 0.1); /* Semi-transparent red */
    color: red;
    display: flex;
    align-items: center;
    justify-content: center;
    pointer-events: none; /* Allow clicks to pass through */
    font-size: 0.875rem;
    padding: 4px;
    box-sizing: border-box;
    word-wrap: break-word;
    overflow: hidden;
    text-overflow: ellipsis;
    z-index: 10000; /* Extremely high z-index */
  }
</style>
<template>
  <IftaLabel>
    <div class="date-picker-container" :class="{ 'has-error': hasError }">
      <DatePicker
        inputId="date"
        id="datepicker-timeonly"
        v-model="selectedTime"
        showIcon
        fluid
        iconDisplay="input"
        timeOnly
        @blur="handleEnterKey"
        :disabled="props.disabled">
        <template #inputicon="slotProps">
          <i class="pi pi-clock" @click="slotProps.clickCallback" />
        </template>
      </DatePicker>
    </div>
    <div v-if="hasError" class="error-overlay">
      <!-- {{ errorMessage }} -->
    </div>
    <label v-if="label" for="date">{{ props.label }}</label>
  </IftaLabel>
</template>
