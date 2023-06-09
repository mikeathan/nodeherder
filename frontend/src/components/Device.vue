<script setup>
import { getSensorIcon } from "../modules/sensors/icons";
import {
  formatLastSeen,
  getBatteryIcon,
  formatLinkQuality,
  getLinkQualityIcon,
} from "../modules/sensors/device";
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
  <div class="card" style="width: 20rem">
    <div class="card-body">
      <h5 class="card-title">{{ device.name }}</h5>

      <div class="card-text" v-for="(value, sensor) in device.payload">
        <div v-if="isSensorWhitelisted(sensor)">
          <div :class="`fa fa-fw ${getSensorIcon(sensor)}`"></div>
          {{ formatSensorValue(sensor, value) }}
        </div>
      </div>
      <div class="card-text">
        <span style="margin-right: 2.5rem">
          {{ formatLastSeen(device.payload) }}
        </span>
        <span
          :class="`fa fa-fw ${getLinkQualityIcon()}`"
          style="margin-right: 0.25rem !important"
        ></span>
        <span style="margin-right: 0.5rem">
          {{ formatLinkQuality(device.payload) }}
        </span>
        <span :class="`fa fa-fw ${getBatteryIcon(device.payload)}`"></span>
      </div>
    </div>
  </div>
</template>
