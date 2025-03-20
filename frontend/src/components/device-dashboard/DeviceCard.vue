<script setup lang="ts">
  import { PropType, computed, ref } from 'vue';
  import DeviceFooter from './DeviceCardFooter.vue';
  import Sensor from '../device/Sensor.vue';
  import { RouterLink } from 'vue-router';
  import Card from 'primevue/card';
  import { Device, Expose } from '@/types/device';

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
        <Sensor :id="device.id" :expose="device.exposes[sensor]" />
      </div>
    </template>
    <template #footer>
      <DeviceFooter :device="device"></DeviceFooter>
    </template>
  </Card>
</template>
