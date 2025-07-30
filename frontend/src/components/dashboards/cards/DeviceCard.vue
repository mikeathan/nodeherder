<script setup lang="ts">
  import { PropType, computed, ref } from 'vue';
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

  const deviceConfig = computed(() => {
    return store.getters['hub/findDeviceSetting'](props.device.id) as DeviceConfig;
  });

  const isDisabled = computed(() => deviceConfig.value?.disabled === true);
  const isOffline = computed(() => !isDeviceOnline(device.value));

  const device = ref<Device>(props.device);
  const measurementExposes = computed(() => {
    return Object.fromEntries(
      Object.entries(props.device.exposes).filter(([key, expose]) => expose.category === 'measurement')
    );
  });
</script>
<style scoped>
  .card-title {
    display: flex;
    justify-content: center;
    align-items: center;
  }
</style>
<template>
  <Card>
    <template #title>
      <div class="card-title">
        <RouterLink :to="`/devicepage/${device.id}`">
          <Button label="Link" variant="link" class="ps-0">
            <h4>{{ device.friendly_name }}</h4>
          </Button>
        </RouterLink>
      </div>
    </template>
    <template #content>
      <DeviceStatusOverlay v-if="isDisabled" :icon="getIconForType('disabled')" text="Device is disabled" rounded />
      <DeviceStatusOverlay v-else-if="isOffline" :icon="getIconForType('offline')" text="Device is offline" rounded />
      <template v-else>
        <div class="flex align-items-center" v-for="(_, sensor) in measurementExposes" :key="sensor">
          <Sensor :id="device.id" :expose="device.exposes[sensor]" :disabled="isOffline" />
        </div>
      </template>
    </template>
    <template v-if="!isDisabled && !isOffline" #footer>
      <DeviceFooter :device="device" />
    </template>
  </Card>
</template>
