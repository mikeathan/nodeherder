<script setup lang="ts">
  import { ref, watchEffect } from 'vue';

  const props = defineProps<{
    message?: string;
    show: boolean;
  }>();

  const emit = defineEmits(['confirm', 'close']);

  function onConfirm() {
    emit('confirm');
    close();
  }

  const showDialog = ref<boolean>(props.show);
  watchEffect(() => (showDialog.value = props.show));

  function close() {
    emit('close', false);
    showDialog.value = false;
  }

  const dialogMessage = () => props.message ?? 'Are you sure?';
</script>

<template>
  <Dialog v-model:visible="showDialog" modal header="Confirm dialog" :style="{ width: '20rem' }">
    <div class="flex items-center gap-4 mb-4">
      <label class="font-semibold w-15">{{ dialogMessage() }}</label>
    </div>
    <div class="flex justify-end gap-2">
      <Button type="button" label="Cancel" severity="secondary" @click="close()"></Button>
      <Button type="button" label="Ok" @click="(e) => onConfirm()"></Button>
    </div>
  </Dialog>
</template>
