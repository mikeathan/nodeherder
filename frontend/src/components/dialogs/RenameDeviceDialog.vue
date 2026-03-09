<script setup lang="ts">
  import { ref, watchEffect, watch } from 'vue';

  const props = defineProps<{
    friendlyName: string;
    show: boolean;
  }>();

  const emit = defineEmits(['confirm', 'close']);

  function rename() {
    emit('confirm', friendlyName.value);
    close();
  }

  const friendlyName = ref<string>('');
  const showDialog = ref<boolean>(props.show);

  watchEffect(
    () => (friendlyName.value = props.friendlyName)
  );
  watchEffect(() => (showDialog.value = props.show));

  function close() {
    emit('close', false);
    showDialog.value = false;
  }
  function isValid() {
    return (
      friendlyName.value != '' &&
      friendlyName.value != props.friendlyName
    );
  }
</script>

<template>
  <Dialog
    v-model:visible="showDialog"
    modal
    header="Rename device"
    :style="{ width: '25rem' }" @hide="close()">
    <div class="flex items-center gap-4 mb-4">
      <label for="friendlyNameId" class="font-semibold w-24"
        >Friendly name</label
      >
      <InputText
        id="friendlyNameId"
        class="flex-auto"
        v-model="friendlyName"
        autocomplete="off" />
    </div>
    <div class="flex justify-end gap-2">
      <Button
        type="button"
        label="Cancel"
        severity="secondary"
        @click="close()"></Button>
      <Button
        type="button"
        label="Save"
        :disabled="isValid() == false"
        @click="(e) => rename()"></Button>
    </div>
  </Dialog>
</template>
