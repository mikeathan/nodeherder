<script setup lang="ts">
  import { ref, watchEffect, computed, onMounted, onUnmounted } from 'vue';
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

  const isMobile = ref(false);

  const checkMobile = () => {
    if (typeof window !== 'undefined') {
      isMobile.value = window.innerWidth <= 640;
    } else {
      isMobile.value = false; // Default for SSR or environments without window
    }
  };

  onMounted(() => {
    checkMobile();
    window.addEventListener('resize', checkMobile);
  });

  onUnmounted(() => {
    window.removeEventListener('resize', checkMobile);
  });

  watchEffect(() => (showDialog.value = props.show));

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
    if (isMobile.value) {
      return {
        width: '100vw',
        height: '100dvh', // Use dynamic viewport height for mobile
        maxHeight: '100dvh', // Match height
        margin: '0',
        transform: 'none',
        borderRadius: '0',
        zIndex: '9999',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      };
    }
    // Default desktop styles (remain the same)
    return {
      width: '30vw',
      minWidth: '300px',
      height: '60vh',
      minHeight: '300px',
      borderRadius: '1rem',
      overflow: 'hidden',
      display: 'flex',
      flexDirection: 'column',
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
      <div class="content-aligner">
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
    </div>
  </Dialog>
</template>
<style scoped>
  .dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    padding: 0.8rem 1rem;
    border-bottom: 1px solid #e9ecef;
    flex-shrink: 0;
  }
  .modal-body {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 1rem;
    box-sizing: border-box;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .content-aligner {
    display: flex;
    flex-direction: column;
    align-items: center;
    width: 100%;
    margin-top: auto; /* THIS is the key change to push it down */
    padding-bottom: 1rem; /* Add some padding at the bottom of the content */
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
