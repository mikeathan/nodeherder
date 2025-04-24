<script setup lang="ts">
  import { getSensorValue, getSensorIcon, getSensorUnit } from '../../modules/formatters/sensor-formatter';
  import { getEntityIcon } from '../../modules/formatters/entity.formatter';
  import { store } from '../../store/index';
  import { computed, ref } from 'vue';
  import { Device, Expose } from '@/types/device';
  import Icon from '../controls/Icon.vue';
  import { ExposeAccessModes, ExposeCategories, ExposeTypes } from '@/types/device.type';
  import { getExposeBinaryProperty, getExposes, toggleExposeBinaryProperty } from '@/contracts/device';
  import { stateDevicesFilter } from '@/configs/automation/device.config';

  // we check if device has state expose and is not the current one
  // if we have we wire the state to icon click
  // if current expose is numeric and writable we show the slider
  // which updates the value
  // but toggle enables/disables the entity

  const props = defineProps({
    id: { type: String, required: true },
    name: { type: String, required: true },
  });

  const device = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    //if (!device) return null;

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

  //if (!device) return null;
  const expose = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    //if (!device) return null;

    return device.exposes[props.name] as Expose;
  });

  const iconProps = computed(() => {
    const icon = getEntityIcon(expose.value.name, expose.value.data);

    if (!isEnabled()) {
      return { ...icon, color: '#9e9e9e' };
    }

    return icon;
  });

  function isReadOnly(): boolean {
    return expose.value.access_mode == ExposeAccessModes.Read;
  }

  function isEnabled() {
    if (device.value.availability == 'offline') {
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
    if (device.value.availability == 'offline') {
      return;
    }
    // check if device has state and toggle

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
  function hasNumericFeatures(): Boolean {
    return expose.value.type == ExposeTypes.Numeric;
  }
  function getUnit() {
    return expose.value.unit == undefined ? getSensorUnit(expose.value.name) : expose.value.unit;
  }
</script>

<template>
  exposedata {{ expose.data }} - statedata {{ stateExpose?.data }} - isEnabled:{{ isEnabled() }}
  <Card class="entity-card">
    <template #title>
      <div class="entity-header">
        <div class="entity-icon" @click="handleIconClick">
          <Icon :icon="iconProps" size="38" background="#363636" />
        </div>
        <div class="entity-labels">
          <div class="entity-title">{{ expose.name }}</div>
          <div class="entity-value">{{ getSensorValue(expose.data) }} {{ getUnit() }}</div>
        </div>
      </div>
    </template>
    <template #content>
      <div v-if="hasNumericFeatures() && !isReadOnly()">
        <Brightness :value="expose.data" @update="updateValue" :min="0" :max="100" :disabled="!isEnabled()" />
      </div>
    </template>
  </Card>
</template>
<style scoped>
  .entity-card {
    border-radius: 8px;

    border: 1px solid wheat;
  }
  .entity-header {
    display: flex;
    align-items: center;
    gap: 0.75em;
  }

  .entity-labels {
    display: flex;
    flex-direction: column;
    justify-content: center;
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
    font-size: 0.75rem;
    color: #777;
    margin-top: 0.2em;
  }

  .entity-icon {
    cursor: pointer;
    transition: background-color 0.2s;
  }
  /* .entity-icon:hover svg path {
    fill: #42a5f5 !important; 
  } */
</style>
