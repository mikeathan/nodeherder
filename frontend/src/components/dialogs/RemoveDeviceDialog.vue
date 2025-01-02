<script setup lang="ts">
  import { ref, watchEffect, onMounted, watch } from 'vue';
  import { prop } from 'vue-class-component';
  import { RemoveDeviceEvent } from '../../types/dialog.type';

  const props = defineProps<{
    friendlyName: string;
    show: boolean;
  }>();

  const emit = defineEmits<{
    (e: 'remove', event: RemoveDeviceEvent): void;
    (e: 'close'): void;
  }>();

  function onRemove() {
    emit('remove', {
      friendlyName: friendlyName.value,
      force: forceRemove.value,
    });
    close();
  }

  const friendlyName = ref<string>('');
  const showDialog = ref<boolean>(props.show);

  const forceRemove = ref<boolean>(false);
  watchEffect(
    () => (friendlyName.value = props.friendlyName)
  );
  watchEffect(() => (showDialog.value = props.show));

  function close() {
    emit('close');
    showDialog.value = false;
  }
</script>

<template>
  <Dialog
    v-model:visible="showDialog"
    modal
    header="Rename device"
    :style="{ width: '25rem' }">
    <div class="flex items-center gap-4 mb-4">
      {{ props.friendlyName }}
    </div>
    <div class="flex items-center gap-4 mb-4">
      <label for="forceRemoveId" class="font-semibold w-24"
        >Force remove</label
      >
      <ToggleSwitch
        id="forceRemoveId"
        v-model="forceRemove" />
    </div>
    <div class="flex justify-end gap-2">
      <Button
        type="button"
        label="Cancel"
        severity="secondary"
        @click="close()"></Button>
      <Button
        type="button"
        label="Remove"
        @click="(e) => onRemove()"></Button>
    </div>
  </Dialog>
</template>
