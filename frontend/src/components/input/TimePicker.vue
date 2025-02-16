<script setup lang="ts">
import { computed, ref } from 'vue';

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
  disabled:{
    type: Boolean,
    default: false,
    required: false,
  }
});

const selectedTime = ref<Date>(props.value);

// const time = computed(() => {
//   if (selectedTime.value) {
//     const tp = toTimePicker(selectedTime.value);
//     const time = new Date();
//     time.setHours(tp.hours);
//     time.setMinutes(tp.minutes);
//     time.setSeconds(0);
//     time.setMilliseconds(0);
//     return time;
//   }
//   return new Date();
// });

const emit = defineEmits<{
  (e: 'updated', value: any): void;
}>();

function handleEnterKey() {
  emit('updated', selectedTime.value);
}
</script>

<template>
  <IftaLabel>

    <DatePicker inputId="date" id="datepicker-timeonly" v-model="selectedTime" showIcon fluid iconDisplay="input"
      timeOnly @blur="handleEnterKey" :disabled="props.disabled">
      <template #inputicon="slotProps">
        <i class="pi pi-clock" @click="slotProps.clickCallback" />
      </template>
    </DatePicker>
    <label v-if="label" for="date">{{ props.label }}</label>

  </IftaLabel>

</template>
