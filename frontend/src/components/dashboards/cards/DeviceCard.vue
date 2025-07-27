<script setup lang="ts">
  import { PropType, computed, ref } from 'vue';
  import DeviceFooter from '../cards/DeviceCardFooter.vue';
  import Sensor from '../../device/Sensor.vue';
  import { RouterLink } from 'vue-router';
  import Card from 'primevue/card';
  import { Device, Expose } from '@/types/device';
  import { isDeviceOnline } from '@/contracts/device';
  import { store } from '@/store';
  import { DeviceConfig } from '@/types/settings.type';

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

  const device = ref<Device>(props.device);
  const measurementExposes = computed(() => {
    return Object.fromEntries(
      Object.entries(props.device.exposes).filter(([key, expose]) => expose.category === 'measurement')
    );
  });
</script>
<style scoped>
  .disabled-card {
    opacity: 0.5;
    filter: grayscale(90%);
    position: relative;
  }
  .disabled-overlay {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding-bottom: 0.5rem;
    text-align: center;
  }
</style>
<template>
  <Card :class="{ 'disabled-card': isDisabled }">
    <template #title>
      <RouterLink :to="`/devicepage/${device.id}`">
        <Button label="Link" variant="link" class="ps-0">
          <h4>{{ device.friendly_name }}</h4>
        </Button>
      </RouterLink>
    </template>
    <template #content>
      <div v-if="isDisabled" class="flex align-items-center disabled-overlay">
        <i class="pi pi-ban" style="font-size: 2rem; color: gray"></i>
        <span>Device is disabled</span>
      </div>
      <template v-else>
        <div class="flex align-items-center" v-for="(_, sensor) in measurementExposes" :key="sensor">
          <Sensor :id="device.id" :expose="device.exposes[sensor]" :disabled="!isDeviceOnline(device)" />
        </div>
      </template>
    </template>
    <template v-if="!isDisabled" #footer>
      <DeviceFooter :device="device" />
    </template>
  </Card>
</template>
