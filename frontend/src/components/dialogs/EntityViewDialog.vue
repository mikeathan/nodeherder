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
    title?: string;
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

  function isToggleable(): boolean {
    return (
      (expose.value.type == ExposeTypes.Binary && expose.value.access_mode != ExposeAccessModes.Read) ||
      stateExpose.value != null
    );
  }
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

  const dialogTitle = () => props.title ?? expose.value.name;
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
    class="p-dialog-custom entity-modal"
    :breakpoints="{ '640px': '100vw' }"
    :style="dialogStyle"
    @hide="close()">
    <template #header>
      <div class="dialog-header">
        <span>{{ dialogTitle() }}</span>
        <Button icon="pi pi-times" class="p-button-text" @click="close()" />
      </div>
    </template>
    <div class="modal-body">
      <div v-if="hasNumericFeatures() && !isReadOnly()">
        <Brightness
          direction="vertical"
          :value="expose.data"
          :min="getExposeAttribute(expose, 'min')"
          :max="getExposeAttribute(expose, 'max')" />
        <div v-if="isToggleable()" class="toggle-wrapper">
          <Button class="toggle-button">
            <i class="pi pi-power-off" style="font-size: 1.3rem" />
          </Button>
        </div>
      </div>
    </div>
  </Dialog>
</template>
<style scoped>
  .dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
  }
  .modal-body {
    flex-grow: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
  }

  .toggle-button {
    width: 3rem;
    height: 3rem;
    border-radius: 50% !important;
    border-color: #222222;
    background-color: #222222;
    padding: 0;
  }

  .toggle-button.p-button:hover {
    box-shadow: none !important;
    background-color: inherit !important;
    border-color: inherit !important;
  }

  .toggle-button i {
    color: white;
  }

  .toggle-wrapper {
    display: flex;
    justify-content: center;
    align-items: center;
    margin-top: 1.5rem;
    width: 100%;
  }
</style>
