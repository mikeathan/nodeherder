<script setup>
import DeviceFooter from './DeviceCardFooter.vue';
import Sensor from '../device/Sensor.vue';
import { RouterLink } from 'vue-router';
import Card from 'primevue/card';

const props = defineProps({
  device: Object,
});
</script>
<style lang="css" scoped></style>
<template>
  <Card :class="device.properties.availability == 'offline' ? 'disabled-card' : ''">
    <template #title>
      <h4>
        <RouterLink :to="`/devicepage/${device.id}`">{{
          device.friendly_name
          }}</RouterLink>
      </h4>
    </template>
    <template #subtitle>

      <p>{{ device.description }}</p>
    </template>
    <template #content>
      <div class="d-flex align-items-center" v-for="(value, sensor) in device.exposes">
          <Sensor :id="device.id" :expose="device.exposes[sensor]" />
        </div>
    </template>
    <template #footer>
      <div class="grid justify-content-between align-items-center">
        <div>
          <i class="pi pi-battery-half" style="font-size: 1.5em;"></i>
          <span>50%</span>
        </div>
        <div>
          <i class="pi pi-wifi" style="font-size: 1.5em;"></i>
          <span>Connected</span>
        </div>
      </div>
    </template>
  </Card>
  <!-- <div class="col-xl-3 col-lg-4 col-sm-6 col-12 d-flex">
    <div class="card flex-fill flex-shrink-1"
      :class="device.properties.availability == 'offline' ? 'disabled-card' : ''">
      <div class="card-header pb-0 d-flex justify-content-left">
        <h4>
          <RouterLink :to="`/devicepage/${device.id}`">{{
            device.friendly_name
            }}</RouterLink>
        </h4>
      </div>
      <div class="card-body row align-items-center">
        <div class="d-flex align-items-center" v-for="(value, sensor) in device.exposes">
          <Sensor :id="device.id" :expose="device.exposes[sensor]" />
        </div>
      </div>
      <DeviceFooter :device="device"></DeviceFooter>
    </div>
  </div> -->
</template>
