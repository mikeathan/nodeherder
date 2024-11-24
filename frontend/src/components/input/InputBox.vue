<script setup lang="ts">
import { ref, watch } from 'vue';

import { InputNumberInputEvent } from 'primevue';

const props = defineProps({
  value: null,
  isNumeric: {
    type: Boolean,
    default: false,
    required: false,
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
});

const emit = defineEmits<{
  (e: 'updated', value: any): void;
  (e: 'lostFocus', value: any): void;
}>();


const inputValue = ref<any>(props.value);
watch(
  () => props.value,
  () => {
    inputValue.value = props.value;
  },
  { immediate: true },
);

function inputNumberEvent(event: InputNumberInputEvent): void {
  const newValue = event.value;
  if (newValue == null || typeof newValue === 'string' || isNaN(newValue)) {
    inputValue.value = 0;
  } else {
    inputValue.value = newValue;
  }
  emit('updated', inputValue.value);
}

function inputTextEvent(event: Event): void {
  let value: any = (event.target as HTMLInputElement).value;
  inputValue.value = value;
  emit('updated', inputValue.value);
}

function onLostFocus(event: Event): void {
  emit('lostFocus', inputValue.value);
}


</script>

<template>
  <FloatLabel variant="in">
    <div v-if="props.isNumeric">
      <InputNumber v-if="props.isNumeric" v-model="inputValue" :disabled="props.disabled" inputId="integeronly"
        @input="inputNumberEvent" :onblur="onLostFocus" :min="0" :max="100" />
      <label v-if="props.label != ''">{{ props.label }}</label>

    </div>
    <div v-else>
      <InputText v-model="inputValue" variant="filled" :disabled="props.disabled" @input="inputTextEvent"
        :onblur="onLostFocus" />
      <label v-if="props.label != ''">{{ props.label }}</label>
    </div>
  </FloatLabel>

</template>
