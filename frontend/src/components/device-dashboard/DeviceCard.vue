<script setup>
import DeviceFooter from "./DeviceCardFooter.vue";
import Availability from "../device/Availability.vue";
import Sensor from "../device/Sensor.vue";
import { RouterLink } from "vue-router";
const props = defineProps({
  device: Object,
});
</script>
<template>
  <div class="col-xl-3 col-lg-4 col-sm-6 col-12 d-flex">

    <div :class="device.stats.availability == 'offline'
      ? 'card flex-fill flex-shrink-1 disabled-card'
      : 'card flex-fill flex-shrink-1'
      ">
      <div class="card-header pb-0 d-flex justify-content-left">

        <div className="me-3">
          <Availability :status="device.stats.availability" />
        </div>
        <h4>
          <RouterLink :to="`/devicepage/${device.id}`">{{
            device.id
          }}</RouterLink>
        </h4>
      </div>
      <div class="card-body row align-items-center">
        <div class="d-flex align-items-center" v-for="(value, sensor) in device.sensors">
          <Sensor :name="sensor" :value="value" />
        </div>
      </div>
      <DeviceFooter :payload="device"></DeviceFooter>
    </div>
  </div>
</template>
