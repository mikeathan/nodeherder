<script setup lang="ts">
  import DebounceEditor from './DebounceEditor.vue';
  import { DeviceDebounce } from '@/types/settings.type';
  import { DefaultExposeCategoryList } from '@/types/device.type';
  import { PropType } from 'vue';
  import { emitOpenSelectionDialog } from '@/contracts/dialog-events';
  import { key } from '@/store';

  const props = defineProps({
    value: { type: Object as PropType<DeviceDebounce>, required: true },
    disabled: { type: Boolean, default: false },
    categories: {
      type: Array as PropType<string[]>,
      default: () => DefaultExposeCategoryList,
    },
  });

  const emit = defineEmits<{
    (e: 'update', value: DeviceDebounce): void;
  }>();

  function handleAddCategory(callback: (key: string) => void) {

    // filter out used keys
    const usedKeys = Object.keys(props.value);
    const availableKeys = props.categories.filter((k) => !usedKeys.includes(k));

    const dlgProps = {
      title: 'Selection',
      message: 'Add new expose category',
      items: availableKeys,
    };
    emitOpenSelectionDialog((key: string) => callback(key), dlgProps);
  }
</script>

<template>
  <DebounceEditor
    :value="value"
    :disabled="disabled"
    :selectable-keys="categories"
    label="Category"
    @update="emit('update', $event)"
    @add-key="handleAddCategory" />
</template>
