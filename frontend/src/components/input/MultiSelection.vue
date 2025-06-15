<script setup lang="ts">
  import { PropType, ref, watch, watchEffect } from 'vue';
  import { SelectSize, SelectFormSize, SelectionItems } from '@/types/controls.type';
  import { MultiSelect, MultiSelectChangeEvent } from 'primevue';

  const props = defineProps({
    items: {
      type: Object as PropType<SelectionItems>,
      default: [],
      required: true,
    },
    values: null,
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
    showClear: {
      type: Boolean,
      default: false,
      required: false,
    },
  });

  const emit = defineEmits<{
    (e: 'updated', value: string[]): void;
  }>();

  const selectedValues = ref<any[]>(props.values);
  const isKeyValuePair = ref<boolean>(false);

  watchEffect(() => {
    if (props.values != null) {
      selectedValues.value = props.values;
    }
  });

  watch(
    () => props.items,
    (newItems) => {
      isKeyValuePair.value = !Array.isArray(newItems) && Object.entries(newItems).length > 0;
    },
    { immediate: true }
  );

  const selectionItems = () => {
    if (Array.isArray(props.items)) {
      return props.items;
    }
    return Object.entries(props.items).map(([key, value]) => ({
      key,
      value,
    }));
  };

  function selectionChanged(event: MultiSelectChangeEvent): void {
    selectedValues.value = event.value;
    emit('updated', selectedValues.value);
  }
</script>

<template>
  <FloatLabel class="w-full md:w-56" variant="on">
    <MultiSelect
      v-model="selectedValues"
      :showClear="props.showClear"
      :options="selectionItems()"
      :optionLabel="isKeyValuePair ? 'key' : ''"
      :optionValue="isKeyValuePair ? 'value' : ''"
      @change="selectionChanged"
      :disabled="props.disabled"
      class="w-full" />
    <label v-if="props.label != ''">{{ props.label }}</label>
  </FloatLabel>
</template>
