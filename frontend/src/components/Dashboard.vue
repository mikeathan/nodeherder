<script setup>
import { getSensorIcon } from "../modules/sensors/icons";
const props = defineProps({
  device: Object,
});

const sensorReadingsWhitelist = {
  "temperature": "Temperature",
  "pressure": "Pressure",
  "humidity": "Humidity",
};

const sensorUnits = {
  temperature: "°C",
  pressure: "$hPa",
  humidity: "%",
  voltage: "mV",
  linkquality: "LQI",
};

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

function getUnit(sensor) {
  if (sensorUnits[sensor] == undefined) {
    return "";
  }
  return sensorUnits[sensor];
}

function format(sensor, value) {
  // icon type unit
  var sensorName = sensor.charAt(0).toUpperCase() + sensor.slice(1);
  return sensorName + " " + value + getUnit(sensor);
}


function isSensorReading(reading) {
  return sensorReadingsWhitelist[reading] != undefined;
}

function isDeviceReading(reading) {
  return deviceReadingsWhitelist[reading] != undefined;
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
