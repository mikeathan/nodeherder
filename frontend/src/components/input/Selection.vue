<script setup lang="ts">
import { KeyValuePair } from '@/types/types';
import { PropType, ref, watch, watchEffect } from 'vue';
import {
  SelectSize,
  SelectFormSize,
  SelectionItems,
  LayoutPosition,
  LayoutPositions,
} from '@/types/controls.type';
import { SelectChangeEvent } from 'primevue';

const props = defineProps({
  items: {
    type: Object as PropType<SelectionItems>,
    default: [],
    required: true,
  },
  value: null,
  text: {
    type: String,
    default: '',
    required: false,
  },
  label: {
    type: String,
    default: '',
    required: false,
  },
  size: {
    type: String as PropType<SelectSize>,
    default: SelectFormSize.small,
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
}>();

const selectedValue = ref<any>(props.value);
const isKeyValuePair = ref<boolean>(false);

watch(
  () => props.items,
  (newItems) => {
    isKeyValuePair.value =
      !Array.isArray(newItems) &&
      Object.entries(newItems).length > 0;
  },
  { immediate: true },
);

const selectionItems = () => {
  if (Array.isArray(props.items)) {
    return props.items;
  }
  return Object.entries(props.items).map(
    ([key, value]) => ({
      key,
      value,
    }),
  );
};

const defaultText = (): string => {
  return props.text != '' ? props.text : 'Select';
};

function selectionChanged(event: SelectChangeEvent): void {
  selectedValue.value = event.value;
  emit('updated', selectedValue.value);
}
</script>

<template>
  <FloatLabel class="w-full md:w-56" variant="on">
    <Select v-model="selectedValue" :options="selectionItems()" :optionLabel="isKeyValuePair ? 'key' : ''"
      :optionValue="isKeyValuePair ? 'value' : ''" @change="selectionChanged" class="w-full" />
    <label v-if="props.label != ''">{{ props.label }}</label>
  </FloatLabel>
</template>
