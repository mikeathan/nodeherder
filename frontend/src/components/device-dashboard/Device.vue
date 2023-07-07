<script setup>
import DeviceFooter from "./DeviceFooter.vue";

import {
  getSensorValue,
  getSensorIcon,
  getSensorName,
} from "../../modules/sensors/sensor-formatter";

const props = defineProps({
  device: Object,
});
</script>
<template>
  <div class="col-xl-3 col-lg-4 col-sm-6 col-12 d-flex">
    <div class="card flex-fill flex-shrink-1">
      <div class="card-header pb-0 d-flex justify-content-between">
        {{ device.id }}
      </div>
      <div
        class="card-body row align-items-center"
        v-for="(value, sensor) in device.payload.sensors"
      >
        <div class="d-flex align-items-center">
          <div class="me-1">
            <i :class="`fa fa-fw ${getSensorIcon(sensor)}`"></i>
          </div>
          <div class="flex-shrink-1 flex-grow-1">
            {{ getSensorName(sensor) }}
          </div>
          <div class="flex-shrink-1">
            {{ getSensorValue(sensor, value) }}
          </div>
        </div>
      </div>
      <p></p>
      <DeviceFooter :payload="device.payload"></DeviceFooter>
    </div>
  </div>
</template>
