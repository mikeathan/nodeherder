<script setup lang="ts">
  import { PropType, computed, ref } from 'vue';
  import DeviceFooter from '../cards/DeviceCardFooter.vue';
  import Sensor from '../../device/Sensor.vue';
  import { RouterLink } from 'vue-router';
  import Card from 'primevue/card';
  import { Device, Expose } from '@/types/device';
  import { isDeviceOnline } from '@/contracts/device';

  const props = defineProps({
    device: {
      type: Object as PropType<Device>,
      default: {} as Device,
    },
  });
  const device = ref<Device>(props.device);
  const measurementExposes = computed(() => {
    return Object.fromEntries(
      Object.entries(props.device.exposes).filter(([key, expose]) => expose.category === 'measurement')
    );
  });
</script>
<!-- <style scoped>
.disabled-card {
  opacity: 0.6;
  pointer-events: none;
  filter: grayscale(80%);
  transition: opacity 0.3s ease, filter 0.3s ease;
  position: relative;
}

.disabled-card::after {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(200, 200, 200, 0.2);
  z-index: 0;
}
</style> -->
<template>
  <Card
    :class="
      device.availability == 'offline' // to fix not working now
        ? 'disabled-card'
        : ''
    ">
    <template #title>
      <RouterLink :to="`/devicepage/${device.id}`">
        <Button label="Link" variant="link" class="ps-0">
          <h4>{{ device.friendly_name }}</h4>
        </Button>
      </RouterLink>
    </template>
    <template #content>
      <div class="flex align-items-center" v-for="(_, sensor) in measurementExposes">
        <Sensor :id="device.id" :expose="device.exposes[sensor]" :disabled="!isDeviceOnline(device)" />
      </div>
    </template>
    <template #footer>
      <DeviceFooter :device="device" />
    </template>
  </Card>
</template>
