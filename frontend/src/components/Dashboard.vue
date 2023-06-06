<script setup>
import { getSensorIcon } from "../modules/sensors/icons";
import { isSensorReading, format } from "../modules/sensors/sensor";
const props = defineProps({
  device: Object,
});


// https://github.com/nurikk/zigbee2mqtt-frontend/blob/dev/src/components/dashboard-page/index.tsx

function formatLastSeen(sensorLastSeen) {
  // find diff between now and last_seen
  var lastSeen = moment(sensorLastSeen);
  var diff = moment().diff(lastSeen);
  var duration = moment.duration(diff);

  // TODO: handle days() > 0
  if (duration.hours() > 0) {
    return duration.hours() + " hours ago";
  }
  if (duration.minutes() > 0) {
    return duration.minutes() + " minutes ago";
  }
  if (duration.seconds() > 5) {
    return duration.seconds() + " seconds ago";
  }

  return "just now";
}

</script>

<template>
  <div class="card" style="width: 18rem">
    <div class="card-body">
      <h5 class="card-title">{{ device.name }}</h5>

      <div class="card-text" v-for="(value, sensor) in device.payload">
        <div v-if="isSensorReading(sensor)">
          <div :class="`fa fa-fw ${getSensorIcon(sensor)}`"></div>
          {{ format(sensor, value) }}
        </div>
      </div>
      <p></p>
      <div class="card-text">Battery and signal stats</div>
    </div>
  </div>
</template>
