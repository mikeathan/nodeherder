<script setup lang="ts">
  import { getFormattedSensorValue, getSensorName } from '../../../modules/formatters/sensor-formatter';
  import { getSensorIcon } from '../../../modules/formatters/sensor-formatter';
  import { store } from '../../../store/index';
  import { computed, nextTick, ref, watch } from 'vue';
  import { Device, Expose } from '@/types/device';
  import Icon from '../../controls/Icon.vue';
  import { ExposeAccessModes, ExposeTypes } from '@/types/device.type';
  import { getExposeBinaryProperty, getExposes, toggleExposeBinaryProperty } from '@/contracts/device';
  import { stateDevicesFilter } from '@/configs/automation/device.config';
  import { emitOpenEntityViewDialog } from '@/contracts/dialog-events';
  import { getDeviceGroupId } from '@/contracts/device-group';
  import DeviceStatusOverlay from '@/components/device/DeviceStatusOverlay.vue';
  import { DeviceConfig } from '@/types/settings.type';
  import { getIconForType } from '@/modules/formatters/icon.formatter';
  import { isDeviceOnline } from '@/contracts/device';
  import { useMiniChartData } from '@/composables/useMiniChartData';
  import { MetricsTypes } from '@/types/metrics.type';
  import SparklineChart from '@/components/chart/mini/SparklineChart.vue';
  import BinarySparklineChart from '@/components/chart/mini/BinarySparklineChart.vue';

  const props = defineProps({
    id: { type: String, required: true },
    name: { type: String, required: true },
    compact: { type: Boolean, required: false, default: false },
    isSelected: { type: Boolean, required: false, default: false },
  });

  const deviceConfig = computed(() => {
    return store.getters['hub/findDeviceSetting'](props.id) as DeviceConfig;
  });

  const device = computed(() => {
    const device = store.getters['hub/findDevice'](props.id) as Device;
    if (!device) return null;

    return device as Device;
  });

  const isDisabled = computed(() => deviceConfig.value?.disabled === true);
  const isOffline = computed(() => device.value && !isDeviceOnline(device.value));
  const showValue = computed(() => !isOffline.value && !isDisabled.value);

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

  // Only fetch chart data if device is online and enabled
  const shouldFetchChart = computed(() => !isDisabled.value && !isOffline.value && device.value != null);
  const { chartData, refetch } = useMiniChartData(props.id, props.name, 24, false);

  // Fetch data when component becomes visible and device is ready
  watch(
    shouldFetchChart,
    (should) => {
      if (should && !chartData.value.hasData && !chartData.value.isLoading) {
        refetch();
      }
    },
    { immediate: true }
  );

  const showMiniChart = computed(() => {
    return (
      chartData.value.hasData &&
      !isDisabled.value &&
      !isOffline.value &&
      (expose.value.type === ExposeTypes.Numeric || expose.value.type === ExposeTypes.Binary)
    );
  });

  const emit = defineEmits<{
    (e: 'delete', value: { id: string; name: string }): void;
    (e: 'selected', id: string): void;
  }>();

  function emitDelete() {
    emit('delete', {
      id: props.id,
      name: props.name,
    });
  }

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
    emit('selected', getDeviceGroupId(props.id, props.name));

    // we need to wait for the next tick to allow consumer to set isSelected
    nextTick(() => {
      if (!props.isSelected) {
        const eventProps = {
          id: props.id,
          name: expose.value.name,
          title: `${getSensorName(expose.value.name)} (${device.value?.friendly_name})`,
        };

        emitOpenEntityViewDialog(() => {}, eventProps);
      }
    });
  }
</script>

<template>
  <Card
    class="entity-card"
    @click="handleCardClick"
    :class="{ 'is-selected': props.isSelected }"
    :pt="{
      body: { style: compact ? 'padding: 0.4rem' : '' },
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
          <div class="entity-title">
            {{ getSensorName(expose.name) }}
          </div>
          <div class="entity-value">
            <template v-if="showValue">{{ getFormattedSensorValue(expose) }}</template>
            <DeviceStatusOverlay v-else-if="isDisabled" :icon="getIconForType('disabled')" :size="18" />
            <DeviceStatusOverlay v-else-if="isOffline" :icon="getIconForType('offline')" :size="18" />
          </div>
        </div>
        <span v-if="isSelected" class="delete-icon pi pi-trash" @click.stop="emitDelete" title="Remove from group" />
      </div>
      <!-- Mini chart background -->
      <SparklineChart
        v-if="showMiniChart && chartData.type === MetricsTypes.Numeric"
        :data="chartData.data as any"
        :exposeName="expose.name"
        :height="60" />
      <BinarySparklineChart
        v-else-if="showMiniChart && chartData.type === MetricsTypes.Binary"
        :data="chartData.data as any"
        :exposeName="expose.name"
        :height="60" />
    </template>
    <template #content> </template>
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
    position: relative;
    overflow: hidden;
  }

  .entity-card:hover {
    background-color: rgba(0, 0, 0, 0.1);
  }

  .entity-card.is-selected {
    border-color: #007bff;
    box-shadow: 0 0 10px rgba(0, 123, 255, 0.5);
    background-color: rgba(0, 123, 255, 0.1);
    cursor: default;
  }
  .entity-header {
    display: flex;
    align-items: flex-end;
    gap: 0.9rem;
    position: relative;
    z-index: 1;
  }

  .entity-labels {
    display: flex;
    flex-direction: column;
    justify-content: center;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
    max-width: 100%;
  }

  .entity-title {
    font-size: 14px;
    font-weight: 500;
  }

  .entity-value {
    font-size: 0.8rem;
    color: #ccc;
    min-height: 1.2rem;
    display: flex;
    gap: 0.25rem;
  }

  .entity-content {
    margin-top: 1rem;
    margin-bottom: 0.3rem;
    margin-left: 0rem;
  }

  .delete-icon {
    position: absolute;
    top: 0;
    right: 0;
    color: #ff5c5c;
    cursor: pointer;
    transition: color 0.2s ease-in-out;
  }
  .delete-icon:hover {
    color: #ff1f1f;
  }
</style>
