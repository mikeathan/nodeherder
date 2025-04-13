<script setup lang="ts">
import { store } from '../../store/index';
import { computed, ref } from 'vue';
import { RemoveDeviceEvent } from '@/types/dialog.type';
import { emitOpenConfirmationDialog, emitOpenDeleteDeviceDialog, emitOpenRenameDeviceDialog } from '@/contracts/dialog-events';

const props = defineProps({
  id: String,
});


const device = computed(() => {
  return store.getters['hub/findDevice'](props.id);
});

function renameDevice(friendlyName: string) {
  if (!friendlyName) {
    console.log('friendlyName is empty');
    return;
  }

  store.dispatch('hub/renameDevice', {
    name: device.value.friendly_name,
    newName: friendlyName,
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


function openRenameDeviceDialog() {
  const props = {
    friendlyName: device.value.friendly_name,
    title: 'Rename Device Friendly Name',
    message: 'Enter new device friendly name',
  }
  emitOpenRenameDeviceDialog((args) => renameDevice(args), props);
}

function openInterviewDeviceDialog() {
  const props = {
    title: 'Interview device',
    message: 'Are you sure?'
  }
  emitOpenConfirmationDialog(interviewDevice, props)
}

function openRemoveDeviceDialog() {
  const props = {
    friendlyName: device.value.friendly_name,
    title: 'Remove device',
    message: 'Are you sure?'
  }
  emitOpenDeleteDeviceDialog((args) => removeDevice(args), props)
}
</script>
<template>

  <Button icon="pi pi-user-edit" variant="text" v-tooltip="'Rename device'" @click="openRenameDeviceDialog" />
  <Button icon="pi pi-sync" variant="text" v-tooltip="'Interview device'" @click="openInterviewDeviceDialog" />
  <Button icon="pi pi-trash" variant="text" v-tooltip="'Remove device'" @click="openRemoveDeviceDialog" />
 
</template>
