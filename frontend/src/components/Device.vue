<script setup>
import { getSensorIcon } from "../modules/sensors/icons";
import {
  isSensorWhitelisted,
  formatSensorValue,
} from "../modules/sensors/sensor";
const props = defineProps({
  device: Object,
});

// https://github.com/nurikk/zigbee2mqtt-frontend/blob/dev/src/components/dashboard-page/index.tsx


</script>

<template>
  <div class="card" style="width: 18rem">
    <div class="card-body">
      <h5 class="card-title">{{ device.name }}</h5>

      <div class="card-text" v-for="(value, sensor) in device.payload">
        <div v-if="isSensorWhitelisted(sensor)">
          <div :class="`fa fa-fw ${getSensorIcon(sensor)}`"></div>
          {{ formatSensorValue(sensor, value) }}
        </div>
      </div>
      <p></p>
      <div class="card-text">Battery and signal stats</div>
    </div>
  </div>
</template>