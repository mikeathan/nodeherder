<script setup lang="ts">
  import { ref, watchEffect, computed, onMounted, onUnmounted } from 'vue';
  import LastSeen from '../device/LastSeen.vue';

  import { store } from '../../store/index';
  import { Device, Expose } from '@/types/device';
  import { getFormattedSensorValue, getSensorName, getSensorUnit } from '../../modules/formatters/sensor-formatter';
  import { ExposeAccessModes, ExposeCategories, ExposeTypes } from '@/types/device.type';
  import {
    getExposeAttribute,
    getExposeBinaryProperty,
    getExposes,
    toggleExposeBinaryProperty,
  } from '@/contracts/device';
  import { stateDevicesFilter, writableExposesDeviceFilter } from '@/configs/automation/device.config';
  import { getEntityIcon } from '@/modules/formatters/entity.formatter';
  import Icon from '../controls/Icon.vue';

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

  const lastSeen = computed(() => {
    return device.value ? device.value.last_seen : '';
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
      isMobile.value = false;
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

  function hasToggle(): boolean {
    return (
      (expose.value.type == ExposeTypes.Binary && expose.value.access_mode != ExposeAccessModes.Read) ||
      stateExpose.value != null
    );
  }

  const controlExposes = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    if (!device) return [];

    const exposeNameList = getExposes(device, writableExposesDeviceFilter());
    return exposeNameList.map((exposeName) => device.exposes[exposeName] as Expose);
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

  const dialogTitle = () => props.title ?? expose.value.name;
  const dialogStyle = computed(() => {
    if (isMobile.value) {
      return {
        width: '100vw',
        height: '100dvh',
        maxHeight: '100dvh',
        margin: '0',
        transform: 'none',
        borderRadius: '0',
        zIndex: '9999',
        display: 'flex',
        flexDirection: 'column',
        overflow: 'hidden',
      };
    }
    return {
      width: '30vw',
      minWidth: '580px',
      maxWidth: '90vw',
      height: 'auto',
      minHeight: '580px',
      maxHeight: '90vh',
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
    :style="dialogStyle"
    @hide="close()">
    <template #header>
      <div class="dialog-header">
        <span>{{ dialogTitle() }}</span>
        <Button icon="pi pi-times" class="p-button-text" @click="close()" />
      </div>
    </template>
    <div class="modal-content-header">
      <div class="modal-value">{{ getFormattedSensorValue(expose) }}</div>
      <LastSeen :timestamp="lastSeen" class="modal-last-seen" />
    </div>

    <div v-if="hasNumericFeatures() && !isReadOnly()" class="modal-content">
      <Brightness
        direction="vertical"
        :value="expose.data"
        :min="getExposeAttribute(expose, 'min')"
        :max="getExposeAttribute(expose, 'max')" />

      <div class="button-panel">
        <template v-for="expose in controlExposes" :key="expose.name">
          <Icon :icon="getEntityIcon(expose.name, expose.data)" background="white" :size="32" :circleRadius="24"/>
        </template>
      </div>
      <div v-if="hasToggle()" class="button-panel">
        <Button class="toggle-button">
          <i class="pi pi-power-off" />
        </Button>
      </div>
    </div>
  </Dialog>
</template>
<style scoped>
  .button-panel {
    display: flex;
    background: #222222;
    border-radius: 999px;
    justify-content: center;
    align-items: center;
    gap: 1rem;
    margin-top: 1.5rem;
    align-self: center;
    width: fit-content;
  }

  .dialog-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    user-select: none;
  }

  .modal-content {
    display: flex;
    flex-direction: column;
    align-items: center;
    user-select: none;
  }

  .modal-content-header {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.2rem;
    padding-bottom: 2rem;
  }

  .modal-value {
    font-size: 36px;
    font-weight: 400;
  }
  .modal-last-seen {
    font-size: 16px;
    font-weight: 500;
  }

  .toggle-wrapper {
    display: flex;
    justify-content: center;
    align-items: center;
    margin-top: 1.5rem;
  }

  .toggle-button {
    width: 3rem;
    height: 3rem;
    border-radius: 50%;
    border-color: #222222;
    background-color: #222222;
    transition: background-color 0.3s ease, border-color 0.3s ease;
  }
  .toggle-button.p-button:hover {
    background-color: #222222 !important;
    border-color: #222222 !important;
    box-shadow: none !important;
  }

  .toggle-button.p-button:active {
    background-color: #4e4e4e !important;
    border-color: #4e4e4e !important;
  }

  .toggle-button i {
    color: white;
    font-size: 1.3rem;
  }
</style>
