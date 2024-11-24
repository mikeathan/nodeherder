<script setup lang="ts">
import { computed, ref } from 'vue';

const props = defineProps({
  value: {
    type: Date,
    default: '',
    required: true,
  },
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
  console.log('handleEnterKey', selectedTime.value);
  emit('updated', selectedTime.value);
}
</script>

<template>
  <DatePicker
    id="datepicker-timeonly"
    v-model="selectedTime"
    showIcon
    fluid
    iconDisplay="input"
    timeOnly
    @blur="handleEnterKey">
    <template #inputicon="slotProps">
      <i
        class="pi pi-clock"
        @click="slotProps.clickCallback" />
    </template>
  </DatePicker>
</template>
