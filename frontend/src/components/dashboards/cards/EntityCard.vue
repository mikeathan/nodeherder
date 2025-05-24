<script setup lang="ts">
  import { getFormattedSensorValue, getSensorName } from '../../../modules/formatters/sensor-formatter';
  import { getSensorIcon } from '../../../modules/formatters/sensor-formatter';
  import { store } from '../../../store/index';
  import { computed, ref } from 'vue';
  import { Device, Expose } from '@/types/device';
  import Icon from '../../controls/Icon.vue';
  import { ExposeAccessModes, ExposeTypes } from '@/types/device.type';
  import { getExposeBinaryProperty, getExposes, toggleExposeBinaryProperty } from '@/contracts/device';
  import { stateDevicesFilter } from '@/configs/automation/device.config';
  import { emitOpenEntityViewDialog } from '@/contracts/dialog-events';

  const props = defineProps({
    id: { type: String, required: true },
    name: { type: String, required: true },
    compact: { type: Boolean, required: false, default: false },
  });

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

  const iconProps = computed(() => {
    const icon = getSensorIcon(expose.value.name, expose.value.data);

    if (!isEnabled()) {
      return { ...icon, color: '#9e9e9e' };
    }

    return icon;
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
    if (stateExpose.value) {
      return getExposeBinaryProperty(stateExpose.value);
    }
    return true;
  }

  function updateValue(exposeName: string, newValue: any): void {
    var msg = {
      id: props.id,
      name: exposeName,
      value: newValue,
    };

    store.dispatch('hub/setDeviceValue', msg);
  }

  function handleIconClick(): void {
    if (device.value?.availability == 'offline') {
      return;
    }

    if (expose.value.type == ExposeTypes.Binary && !isReadOnly()) {
      const value = toggleExposeBinaryProperty(expose.value);
      updateValue(expose.value.name, value);
    } else {
      if (stateExpose.value?.data) {
        const value = toggleExposeBinaryProperty(stateExpose.value);
        updateValue(stateExpose.value.name, value);
      }
    }
  }

  function isToggleable(): boolean {
    return (
      (expose.value.type == ExposeTypes.Binary && expose.value.access_mode != ExposeAccessModes.Read) ||
      stateExpose.value != null
    );
  }

  function handleCardClick(): void {
    const eventProps = {
      id: props.id,
      name: expose.value.name,
      title: getSensorName(expose.value.name),
    };

    // TODO:
    ///send title name to include dashboardgroupName + entity name
    emitOpenEntityViewDialog(() => {}, eventProps);
  }

  const cardStyle = computed(() => {
    if (props.compact) {
      return 'padding: 0.4rem';
    }
    return '';
  });
</script>

<template>
  <Card
    class="entity-card"
    @click="handleCardClick"
    :pt="{
      body: { style: cardStyle },
      root: { style: { '--p-card-body-gap': '0.0rem' } }, // remove card padding
    }">
    <template #title>
      <div class="entity-header">
        <Icon
          :icon="iconProps"
          :size="38"
          :circle-radius="18"
          background="#363636"
          :clickable="isToggleable()"
          @click="handleIconClick" />
        <div class="entity-labels">
          <div class="entity-title">{{ getSensorName(expose.name) }}</div>
          <div class="entity-value">{{ getFormattedSensorValue(expose) }}</div>
        </div>
      </div>
    </template>
    <template #content>
      <!-- <div v-if="hasNumericFeatures() && !isReadOnly()" class="entity-content">
        <Brightness
          :value="expose.data"
          @update="updateValue(expose.name, $event)"
          :min="getExposeAttribute(expose, 'min')"
          :max="getExposeAttribute(expose, 'max')"
          :disabled="!isEnabled()" />
      </div> -->
    </template>
  </Card>
</template>
<style scoped>
  .entity-card {
    cursor: pointer;
    user-select: none;

    width: 100%;
    height: auto;

    min-height: 2rem;
    max-height: 10rem;
    /* max-width: 240px; */

    margin: 0 auto;
    border: 1px solid rgba(0, 0, 0, 0.38);
    border-radius: 12px;
    -webkit-user-select: none; /* Safari */
    -moz-user-select: none; /* Firefox */
    -ms-user-select: none; /* Internet Explorer/Edge */
  }

  .entity-card:hover {
    background-color: rgba(0, 0, 0, 0.1);
  }

  .entity-header {
    display: flex;
    align-items: flex-end;
    gap: 0.9rem;
  }

  .entity-labels {
    display: flex;
    flex-direction: column;
    justify-content: center;
  }

  .entity-title {
    font-size: 1rem;
    font-weight: 500;
  }

  .entity-value {
    font-size: 0.8rem;
    color: #ccc;
    min-height: 1.2rem;
  }

  .entity-content {
    margin-top: 1rem;
    margin-bottom: 0.3rem;
    margin-left: 0rem;
  }
</style>
