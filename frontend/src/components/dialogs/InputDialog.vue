<script setup lang="ts">
  import { ref, watchEffect } from 'vue';

  const props = defineProps<{
    title?: string;
    message?: string;
    show: boolean;
    value?: string;
  }>();

  const emit = defineEmits<{
    (e: 'confirm', value: string): void;
    (e: 'close'): void;
  }>();

  const inputValue = ref(props.value ?? '');
  const showDialog = ref<boolean>(props.show);

  // Watch for prop changes
  watchEffect(() => {
    showDialog.value = props.show;
    inputValue.value = props.value ?? '';
  });

  function confirm() {
    emit('confirm', inputValue.value);
    close();
  }

  function close() {
    emit('close');
    showDialog.value = false;
  }

  const dialogTitle = () => props.title ?? 'Input';
  const dialogMessage = () => props.message ?? '';
  const isValid = () => inputValue.value.length > 0 && inputValue.value != props.value;
</script>

<template>
  <Dialog v-model:visible="showDialog" modal :header="dialogTitle()" :style="{ width: '25rem' }" @hide="close()">
    <div v-if="dialogMessage()" class="mb-3 text-sm text-color-secondary">
      {{ dialogMessage() }}
    </div>

    <InputText v-model="inputValue" class="w-full mb-4" />

    <div class="flex justify-end gap-2">
      <Button type="button" label="Cancel" severity="secondary" @click="close()" />
      <Button type="button" label="OK" @click="confirm()" :disabled="!isValid()" />
    </div>
  </Dialog>
</template>
