<script setup>
import DeviceFooter from './DeviceCardFooter.vue';
import Sensor from '../device/Sensor.vue';
import { RouterLink } from 'vue-router';

const props = defineProps({
  device: Object,
});
</script>
<style scoped>
.sensor-item {
  white-space: nowrap; /* prevent text wrapping */
  overflow: hidden; /* hide overflow */
  text-overflow: ellipsis; /* add ellipsis on overflow */
  padding-top: 5px;
}
</style>
<template>
  <v-card
    class="d-flex flex-column flex-grow-1 w-100 h-100"
    :class="
      device.properties.availability == 'offline'
        ? 'disabled-card'
        : ''
    "
    outlined>
    <v-card-title>
      <RouterLink :to="`/devicepage/${device.id}`">{{
        device.friendly_name
      }}</RouterLink>
    </v-card-title>
    <v-card-text
      class="flex-grow-1 d-flex flex-column justify-content-start">
      <div
        v-for="(value, sensor) in device.exposes"
        :key="sensor"
        class="d-flex align-center sensor-item">
        <Sensor :id="device.id" :expose="value" />
      </div>
    </v-card-text>
    <div
      class="d-flex flex-column flex-grow-1 justify-content-end ps-3 pb-3">
      <DeviceFooter :device="device"></DeviceFooter>
    </div>
  </v-card>
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
