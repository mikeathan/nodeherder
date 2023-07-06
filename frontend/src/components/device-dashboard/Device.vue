<script setup>
import DeviceFooter from "./DeviceFooter.vue";
import "../../assets/css/device.styles.css";

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
  <div className="col-xl-3 col-lg-4 col-sm-6 col-12 d-flex">
    <div className="card flex-fill flex-shrink-1">
      <div class="card-title">{{ device.id }}</div>
      <div class="card-text">
        <div
          class="card-text row align-items-center"
          v-for="(value, sensor) in device.payload.sensors"
        >
          <div class="d-flex align-items-center">
            <div className="me-1">
              <i :class="`fa fa-fw ${getSensorIcon(sensor)}`"></i>
            </div>
            <div className="flex-shrink-1 flex-grow-1">
              <span class="sensor-name-span">{{ getSensorName(sensor) }}</span>
              <span>{{ getSensorValue(sensor, value) }}</span>
            </div>
          </div>
        </div>
      </div>
      <p></p>
      <DeviceFooter :payload="device.payload"></DeviceFooter>
    </div>
  </div>
</template>
