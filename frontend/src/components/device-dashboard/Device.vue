<script setup>
import DeviceFooter from "./DeviceFooter.vue";
import "../../assets/css/device.styles.css";

import {
  isSensorWhitelisted,
  getSensorValue,
  getSensorIcon,
  getSensorName,
} from "../../modules/sensors/sensor";

const props = defineProps({
  device: Object,
});
</script>

<template>
  <div class="card" style="width: 25rem">
    <div class="card-body">
      <h5 class="card-title">{{ device.name }}</h5>

      <div class="card-text" v-for="(value, sensor) in device.payload">
        <div v-if="isSensorWhitelisted(sensor)">
          <span :class="`fa fa-fw ${getSensorIcon(sensor)}`"></span>
          <span class="sensor-name-span">{{ getSensorName(sensor) }}</span>
          <span>{{ getSensorValue(sensor, value) }}</span>
        </div>
      </div>
      <DeviceFooter :payload="device.payload"></DeviceFooter>
    </div>
  </div>
</template>
