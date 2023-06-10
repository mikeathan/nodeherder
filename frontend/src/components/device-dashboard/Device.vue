<script setup>
import DeviceFooter from "./DeviceFooter.vue";

import {
  isSensorWhitelisted,
  getSensorValue,
  getSensorIcon,
  getSensorName
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
          <span style="display: inline-block;overflow: hidden;width:10em;">{{ getSensorName(sensor) }}</span>
          <span>{{ getSensorValue(sensor, value) }}</span>
        </div>
      </div>
      <DeviceFooter :payload="device.payload"></DeviceFooter>
    </div>
  </div>
</template>
