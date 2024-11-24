<script setup lang="ts">
import { ref, watch, PropType } from 'vue';
import {
  SelectSize,
  SelectFormSize,
  SelectionItems,
  LayoutPosition,
  LayoutPositions,
} from '@/types/controls.type';

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
  position: {
    type: String as PropType<LayoutPosition>,
    default: LayoutPositions.left,
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

const inputId = ref(
  `username-${Math.random().toString(36).substring(2, 9)}`,
);
const inputValue = ref<any>(props.value);
watch(
  () => props.value,
  () => {
    inputValue.value = props.value;
  },
  { immediate: true },
);

function inputChanged(event: Event) {
  let value: any = (event.target as HTMLInputElement).value;
  if (props.isNumeric) {
    if ((value as string).endsWith('.')) {
      return;
    }
    value = Number(value);
  }
  inputValue.value = value;
  emit('updated', inputValue.value);
}

function onLostFocus(event: Event) {
  emit('lostFocus', inputValue.value);
}

function isNumber(event: KeyboardEvent) {
  if (
    props.isNumeric &&
    (!event.key.match(/^[\d\.]$/) ||
      isNaN(Number(inputValue.value)))
  ) {
    event.preventDefault();
  }
}
</script>

<template>
  <FloatLabel variant="in">
    <InputText
      :id="inputId"
      v-model="inputValue"
      variant="filled"
      @input="inputChanged"
      @keypress="isNumber"
      :disabled="props.disabled" />
    <label  v-if="props.label != ''" :for="inputId">{{ props.label }}</label>
  </FloatLabel>
  
</template>
