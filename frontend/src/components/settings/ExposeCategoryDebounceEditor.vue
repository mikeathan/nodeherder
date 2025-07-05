<script setup lang="ts">
  import DebounceEditor from './DebounceEditor.vue';
  import { DeviceDebounce } from '@/types/settings.type';
  import { DefaultExposeCategoryList } from '@/types/device.type';
  import { PropType } from 'vue';

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

  function handleAddCategory(cb: (key: string) => void) {
    const cat = prompt('Enter category name', '');
    if (cat && !props.value[cat]) {
      cb(cat);
    }
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
