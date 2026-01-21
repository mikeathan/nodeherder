<script setup lang="ts">
  import { PropType, computed } from 'vue';
  import DeviceFooter from '../cards/DeviceCardFooter.vue';
  import Sensor from '../../device/Sensor.vue';
  import { RouterLink } from 'vue-router';
  import Card from 'primevue/card';
  import { Device } from '@/types/device';
  import { isDeviceOnline } from '@/contracts/device';
  import { store } from '@/store';
  import { DeviceConfig } from '@/types/settings.type';
  import DeviceStatusOverlay from '@/components/device/DeviceStatusOverlay.vue';
  import { getIconForType } from '@/modules/formatters/icon.formatter';

  const props = defineProps({
    device: {
      type: Object as PropType<Device>,
      default: {} as Device,
    },
  });

  const deviceConfig = computed(() => store.getters['hub/findDeviceSetting'](props.device.id) as DeviceConfig);

  const isDisabled = computed(() => deviceConfig.value?.disabled === true);
  const isOffline = computed(() => !isDeviceOnline(props.device));
  const measurementExposes = computed(() => {
    return Object.fromEntries(
      Object.entries(props.device.exposes).filter(([key, expose]) => expose.category === 'measurement')
    );
  });
</script>
<style scoped>
  .device-card {
    border: 1px solid rgba(0, 0, 0, 0.38);
    border-radius: 12px;
    box-shadow: none !important;
    overflow: hidden;
  }
  .card-title {
    display: flex;
    justify-content: center;
    align-items: center;
  }
  h4 {
    margin: 0;
    font-size: 1.1rem;
    font-weight: 500;
  }
</style>
<template>
  <Card class="device-card">
    <template #title>
      <div class="card-title">
        <RouterLink :to="`/devicepage/${props.device.id}`">
          <Button label="Link" variant="link" class="ps-0" text>
            <h4>{{ props.device.friendly_name }}</h4>
          </Button>
        </RouterLink>
      </div>
    </template>
    <template #content>
      <DeviceStatusOverlay v-if="isDisabled" :icon="getIconForType('disabled')" text="Device is disabled" rounded />
      <DeviceStatusOverlay v-else-if="isOffline" :icon="getIconForType('offline')" text="Device is offline" rounded />
      <template v-else>
        <div class="flex align-items-center" v-for="(_, sensor) in measurementExposes" :key="sensor">
          <Sensor :id="props.device.id" :expose="props.device.exposes[sensor]" :disabled="isOffline" />
        </div>
      </template>
    </template>
    <template v-if="!isDisabled && !isOffline" #footer>
      <DeviceFooter :device="props.device" />
    </template>
  </Card>
</template>
