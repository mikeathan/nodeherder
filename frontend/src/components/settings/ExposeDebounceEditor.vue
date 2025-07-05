<script setup lang="ts">
import DebounceEditor from './DebounceEditor.vue';
import { emitOpenExposeSelectionDialog } from '@/contracts/dialog-events';
import { DeviceDebounce } from '@/types/settings.type';

const props = defineProps<{
  id: string;
  value: DeviceDebounce;
  disabled?: boolean;
}>();

const emit = defineEmits<{
  (e: 'update', value: DeviceDebounce): void;
}>();

function handleAddExpose(cb: (key: string) => void) {
  const dlgProps = {
    id: props.id,
    title: 'Select Expose',
    message: 'Select Expose',
  };
  emitOpenExposeSelectionDialog((key: string) => cb(key), dlgProps);
}
</script>

<template>
  <DebounceEditor
    :value="value"
    :disabled="disabled"
    :selectable-keys="Object.keys(value)"
    label="Expose"
    @update="emit('update', $event)"
    @add-key="handleAddExpose"
  />
</template>
