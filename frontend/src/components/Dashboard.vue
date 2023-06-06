<script setup>
const props = defineProps({
  device: Object,
});

const iconPrefix = "fa fa-fw ";
const deviceReadingsWhitelist = {
  voltage: "mV",
  state: " .", // this is just for testing - remove
  battery: "%", // icon and caption
  last_seen: "(WIP)", // 22 minutes ago
  linkquality: "LQI", // icon and caption
};

const getTemperatureIcon = (temperature) => {
  let icon = "fa-thermometer-empty";
  if (temperature >= 30) {
    icon = "fa-thermometer-full";
  } else if (temperature >= 25) {
    icon = "fa-thermometer-three-quarters";
  } else if (temperature >= 20) {
    icon = "fa-thermometer-half";
  } else if (temperature >= 15) {
    icon = "fa-thermometer-quarter";
  }
  return icon;
};
const typeToClassMap = {
  temperature: "fa-thermometer-full",
  humidity: "fa-tint",
  illuminance: "fa-sun",
  pressure: "fa-cloud-download-alt",
  co2: "text-warning",
  voltage: "text-success",
  state: "fa-star-half-alt",
  brightness: "fa-sun",
  occupancy: "fa-walking",
  current: "fa-copyright",
  power: "fa-power-off",
  energy: "fa-plug",
  frequency: "fa-wave-square",
  tamper: "fa-exclamation-circle",
  smoke: "fa-smoking",
  radiation_dose_per_hour: "fa-radiation",
  radioactive_events_per_minute: "fa-radiation-alt",
  power_factor: "fa-industry",
  mode: "fa-user-cog",
  sound: "fa-volume-up",
  position: "fa-percent",
  alarm: "fa-exclamation-triangle",
  color_xy: "fa-palette",
  color_hs: "fa-palette",
  color_temp: "fa-sliders-h",
  illuminance_lux: "fa-sun",
  soil_moisture: "fa-fill-drip",
  water_leak: "fa-water",
  week: "fa-calendar-week",
  workdays_schedule: "fa-calendar-day",
  holidays_schedule: "fa-calendar-day",
  away_mode: "fa-plane",
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

function getIcon(sensor, value) {
  //https://github.com/nurikk/zigbee2mqtt-frontend/blob/dev/src/components/dashboard-page/DashboardFeatureWrapper.tsx
  switch (name) {
    case "device_temperature":
    case "temperature":
    case "local_temperature":
      typeToClassMap[sensor] = getTemperatureIcon(value);
      break;
  }

  return iconPrefix + typeToClassMap[sensor];
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
          <div :class="`fa fa-fw ${getIcon(sensor)}`"></div>
          {{ format(sensor, value) }}
        </div>
      </div>
      <p></p>
      <div class="card-text">Battery and signal stats</div>
    </div>
  </div>
</template>
