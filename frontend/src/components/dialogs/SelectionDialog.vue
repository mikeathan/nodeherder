<script setup lang="ts">
  import { ref, watchEffect } from 'vue';

  TODO;

  const props = defineProps<{
    title?: string;
    message?: string;
    show: boolean;
    value?: string;
    items: string[];
  }>();

  const emit = defineEmits<{
    (e: 'confirm', value: string): void;
    (e: 'close'): void;
  }>();

  const selected = ref<string | null>(null);

  const inputValue = ref(props.value ?? '');
  const showDialog = ref<boolean>(props.show);

  // Watch for prop changes
  watchEffect(() => {
    showDialog.value = props.show;
    inputValue.value = props.value ?? '';
  });

  function select() {
  emit('confirm', selected.value);
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
    <div class="flex items-center gap-4 mb-4">
      <Selection :value="value" :items="items" @updated="(v: any) => { selected = v }" />
    </div>
    <div class="flex justify-end gap-2">
      <Button type="button" label="Cancel" severity="secondary" @click="close()"></Button>
      <Button type="button" label="Save" :disabled="isValid() == false" @click="select()"></Button>
    </div>
  </Dialog>
</template>
