<script setup lang="ts">
  import { store } from '../../store/index';
  import { computed, ref } from 'vue';
  import RenameDeviceDialog from '../dialogs/RenameDeviceDialog.vue';
  import ConfirmDialog from '../dialogs/ConfirmDialog.vue';
  import RemoveDeviceDialog from '../dialogs/RemoveDeviceDialog.vue';
  import { RemoveDeviceEvent } from '@/types/dialog.type';

  const props = defineProps({
    id: String,
  });

  const showRenameDialog = ref(false);
  const showInterviewDialog = ref(false);
  const showRemoveDialog = ref(false);

  const device = computed(() => {
    return store.getters['hub/findDevice'](props.id);
  });

  function renameDevice(value: string) {
    store.dispatch('hub/renameDevice', {
      name: device.value.friendly_name,
      newName: value,
    });
  }

  function interviewDevice() {
    store.dispatch('hub/interviewDevice', {
      id: device.value.id,
    });
  }

  function removeDevice(event: RemoveDeviceEvent) {
    store.dispatch('hub/removeDevice', {
      id: device.value.id,
      force: event.force ?? false,
      block: event.block ?? false,
    });
  }
  
</script>
<template>
 
  <Button icon="pi pi-user-edit" variant="text" v-tooltip="'Rename device'" @click="showRenameDialog = true" />
  <RenameDeviceDialog
    :friendlyName="device.friendly_name"
    :show="showRenameDialog"
    @update:name="renameDevice"
    @close="showRenameDialog = false" />

  <Button icon="pi pi-sync" variant="text" v-tooltip="'Interview device'" @click="showInterviewDialog = true" />
  <ConfirmDialog :show="showInterviewDialog" @confirm="interviewDevice" @close="showInterviewDialog = false" />

  <Button icon="pi pi-trash" variant="text" v-tooltip="'Remove device'" @click="showRemoveDialog = true" />
  <RemoveDeviceDialog
    :friendlyName="device.friendly_name"
    :show="showRemoveDialog"
    @remove="(e) => removeDevice(e)"
    @close="showRemoveDialog = false" />
</template>
