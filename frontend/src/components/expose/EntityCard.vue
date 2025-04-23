<script setup lang="ts">
  import { getSensorValue, getSensorIcon, getSensorUnit } from '../../modules/formatters/sensor-formatter';
  import { getEntityIcon } from '../../modules/formatters/entity.formatter';
  import { store } from '../../store/index';
  import { computed, ref } from 'vue';
  import { Device, Expose } from '@/types/device';
  import Icon from '../controls/Icon.vue';
  import { mdiCeilingLightMultiple } from '@mdi/js';
  import { ExposeAccessModes, ExposeCategories, ExposeTypes } from '@/types/device.type';
  import { getExposes } from '@/contracts/device';
  import { featureDevicesFilter } from '@/configs/automation/device.config';

  const props = defineProps({
    id: { type: String, required: true },
    name: { type: String, required: true },
  });
  const device = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    //if (!device) return null;

    return device as Device;
  });

  const expose = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    //if (!device) return null;

    return device.exposes[props.name] as Expose;
  });

  // const exposeValue = computed(() => {
  //   if (expose.value?.data) {
  //     return getSensorValue(expose.value.data);
  //   }
  //   return null;
  // });

  const iconProps = computed(() => {
    const icon = getEntityIcon(expose.value.name, expose.value.data);

    if (isDisabled()) {
      return { ...icon, color: '#9e9e9e' };
    }

    return icon;
  });

  function isReadOnly(): boolean {
    return expose.value.access_mode == ExposeAccessModes.Read;
  }

  function isDisabled() {
    if (device.value.availability == 'offline') {
      return true;
    }

    if (expose.value.type == ExposeTypes.Numeric && expose.value.data == 0) {
      return true;
    }
    if (expose.value.type == ExposeTypes.Binary && expose.value.data == false) {
      return true;
    }
  }
  function updateValue(newValue: any): void {
    var msg = {
      id: props.id,
      name: expose.value.name,
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
      updateValue(!expose.value.data);
    } else {
      // we can do that but how do we check if now the state is off the brightness is 0?
      // maybe check first for the stae here if entty has this configuration and that drives the control

      ???
      var filtered = getExposes(
        device.value,
        (device: Device, expose: Expose): boolean =>
          expose.type == ExposeTypes.Binary &&
          expose.access_mode != ExposeAccessModes.Read &&
          expose.category == ExposeCategories.Measurement
      );
      console.log(filtered);
      if (filtered.length > 0) {
        const toggle = filtered[0];
        var msg = {
          id: props.id,
          name: toggle,
          value: !device.value.exposes[toggle].data,
        };

        store.dispatch('hub/setDeviceValue', msg);
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
<template>
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
        <Brightness :value="expose.data" @update="updateValue" :min="0" :max="100" />
      </div>
    </template>
  </Card>
</template>
