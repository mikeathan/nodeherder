<script setup lang="ts">
  import { ref, watchEffect, computed, onMounted, onUnmounted, h } from 'vue';
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
  import { writableExposesDeviceFilter } from '@/configs/automation/device.config';
  import { getEntityIcon } from '@/modules/formatters/entity.formatter';
  import Icon from '../controls/Icon.vue';
  import StyledSlider from '../input/StyledSlider.vue';

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

  // todo get Effects from device
  // maybe effects with color temp are at the buttom ofthe button panel?
  // or color temp is shown in brigthness slider with different range and background color of orange
  // and a side info when dragged
  //also maybe we dont show brigthness in the entity card and we only show it in the modal
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

  function isEnabled() {
    if (device.value?.availability == 'offline') {
      return false;
    }

    if (expose.value.type == ExposeTypes.Numeric && expose.value.data == 0) {
      return false;
    }

    if (expose.value.type == ExposeTypes.Binary) {
      return getExposeBinaryProperty(expose.value);
    }

    if (controlExposes.value) {
      // temporary fix for state control
      return controlExposes.value.some((expose) => {
        if (expose.type == ExposeTypes.Binary) {
          return getExposeBinaryProperty(expose);
        }
      });
    }
    return true;
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

  function handleClick(expose: Expose) {
    if (device.value?.availability == 'offline') {
      return;
    }
    if (expose.type == ExposeTypes.Binary) {
      const value = toggleExposeBinaryProperty(expose);
      updateValue(expose.name, value);
    } else {
      console.log('not handled');
    }
  }

  function updateValue(exposeName: string, newValue: any): void {
    var msg = {
      id: props.id,
      name: exposeName,
      value: newValue,
    };

    store.dispatch('hub/setDeviceValue', msg);
  }

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
      <StyledSlider
        direction="vertical"
        :value="expose.data"
        @update="updateValue(expose.name, $event)"
        :min="getExposeAttribute(expose, 'min')"
        :max="getExposeAttribute(expose, 'max')"
        :disabled="!isEnabled()" />

      <div class="button-panel">
        <template v-for="expose in controlExposes" :key="expose.name">
          {{expose.name}} -{{ expose.access_mode }} - {{ expose.type }} - {{expose.values}}
          <Icon
            :icon="getEntityIcon(expose.name, expose.data)"
            clickable
            background="#222222"
            :size="38"
            @click="handleClick(expose)" />
        </template>
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
</style>
