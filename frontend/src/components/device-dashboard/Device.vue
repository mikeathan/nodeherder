<script setup>
import DeviceFooter from "./DeviceFooter.vue";
import { getSensorIcon } from "../../modules/sensors/icons";

import {
  isSensorWhitelisted,
  formatSensorValue,
} from "../../modules/sensors/sensor";

const props = defineProps({
  device: Object,
});
</script>

<template>
  <div class="card" style="width: 20rem">
    <div class="card-body">
      <h5 class="card-title">{{ device.name }}</h5>

      <div class="card-text" v-for="(value, sensor) in device.payload">
        <div v-if="isSensorWhitelisted(sensor)">
          <div :class="`fa fa-fw ${getSensorIcon(sensor)}`"></div>
          {{ formatSensorValue(sensor, value) }}
        </div>
      </div>
      <DeviceFooter :payload="device.payload"></DeviceFooter>
    </div>
  </div>
</template>
