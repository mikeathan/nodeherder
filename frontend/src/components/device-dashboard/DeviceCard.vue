<script setup>
import DeviceFooter from './DeviceCardFooter.vue';
import Sensor from '../device/Sensor.vue';
import { RouterLink } from 'vue-router';
import Card from 'primevue/card';

const props = defineProps({
  device: Object,
});
</script>


<template>
  <Card :class="device.properties.availability == 'offline' ? 'disabled-card' : ''">
    <template #title>
      <RouterLink :to="`/devicepage/${device.id}`">
          <Button label="Link" variant="link" class="ps-0">
            <h4>{{ device.friendly_name }}</h4>
          </Button>
      </RouterLink>
    </template>
    <template #subtitle>
      <p>{{ device.description }}</p>
    </template>
    <template #content>
      <div class="flex align-items-center" v-for="(value, sensor) in device.exposes">
        <Sensor :id="device.id" :expose="device.exposes[sensor]" />
      </div>
    </template>
    <template #footer>
      <DeviceFooter :device="device"></DeviceFooter>
    </template>
  </Card>
</template>
