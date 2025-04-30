<script setup lang="ts">
  import { ref, watchEffect, computed } from 'vue';
  import { store } from '../../store/index';
  import { Device, Expose } from '@/types/device';
  import Selection from '@/components/input/Selection.vue';
  import { ExposeAccessModes, ExposeCategories, ExposeTypes } from '@/types/device.type';
  import {
    getExposeAttribute,
    getExposeBinaryProperty,
    getExposes,
    toggleExposeBinaryProperty,
  } from '@/contracts/device';
  import { stateDevicesFilter } from '@/configs/automation/device.config';

  const props = defineProps<{
    show: boolean;
    id: string;
    name: string;
  }>();

  const emit = defineEmits(['close']);

  const showDialog = ref<boolean>(props.show);

  watchEffect(() => (showDialog.value = props.show));

  const device = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    if (!device) return null;

    return device as Device;
  });

  const stateExpose = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    if (!device) return null;
    var stateExposes = getExposes(device, stateDevicesFilter());

    // for now we only support one state expose
    if (stateExposes.length > 0) {
      if (stateExposes.length > 1) {
        console.error('More than one state expose found: ', stateExposes);
      }
      return device.exposes[stateExposes[0]] as Expose;
    }

    return null;
  });

  const expose = computed(() => {
    if (!device.value) return {} as Expose;

    return device.value.exposes[props.name] as Expose;
  });
  function isReadOnly(): boolean {
    return expose.value.access_mode == ExposeAccessModes.Read;
  }

  function close() {
    emit('close', false);
    showDialog.value = false;
  }

  const dialogTitle = () => 'Title';
  const dialogMessage = () => 'Message';
  const dialogStyle = computed(() => {
    // Mobile screen styling
    if (window.innerWidth <= 640) {
      return {
        width: '100vw',
        height: '100vh',
        maxHeight: '100vh',
        margin: '0',
        padding: '0',
        transform: 'none',
        borderRadius: '0',
        zIndex: '9999',
      };
    }
    // Default desktop styles
    return {
      width: '30vw',
      height: '60vh',
      borderRadius: '1rem',
      overflow: 'auto',
    };
  });

  function hasNumericFeatures(): boolean {
    return expose.value.type == ExposeTypes.Numeric;
  }
</script>

<template>
  <Dialog
    v-model:visible="showDialog"
    :draggable="false"
    :dismissableMask="true"
    :blockScroll="true"
    :closable="false"
    modal
    :header="dialogTitle()"
    class="p-dialog-custom entity-modal"
    :breakpoints="{ '640px': '100vw' }"
    :style="dialogStyle"
    @hide="close()">
    <template #header>
      <div class="dialog-header">
        <span>My Custom Header</span>
        <Button icon="pi pi-times" class="p-button-text" @click="close()" />
      </div>
    </template>
    <div class="modal-body">
      <p><strong>Name:</strong> {{ dialogMessage() }}</p>

      <div v-if="hasNumericFeatures() && !isReadOnly()" class="rotate-wrapper ">
        <Brightness
          :value="expose.data"
        
          :min="getExposeAttribute(expose, 'min')"
          :max="getExposeAttribute(expose, 'max')"
     />
      </div>
    </div>
  </Dialog>
</template>
<style scoped>

.rotate-wrapper {
  display: inline-block;
  transform: rotate(90deg);
  transform-origin: center center;
  width: fit-content;
  height: fit-content;
  overflow: visible;
}

  .dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
  }
</style>
