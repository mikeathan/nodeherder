<script setup>
import DeviceFooter from "./DeviceCardFooter.vue";
import Sensor from "../device/Sensor.vue";
import { RouterLink } from "vue-router";

const props = defineProps({
  device: Object,
});

</script>
<template>
  <div class="col-xl-3 col-lg-4 col-sm-6 col-12 d-flex">
    <div class="card flex-fill flex-shrink-1" :class="device.properties.availability == 'offline' ? 'disabled-card' : ''">
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
  </div>
</template>
