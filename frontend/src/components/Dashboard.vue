<script setup>


const props = defineProps({
  device: Object
})

const extaViewSensorReadings = {
  "voltage": "mV",
  "state":" ."
}

const quickViewSensorReadings= {
  "temperature": "°C",
  "pressure": "hPa",
  "humidity": "%",
  "battery": "%",       // icon and caption
  "last_seen": "(WIP)", // 22 minutes ago
  "linkquality": "LQI", // icon and caption
};

// TODO:
const test= {
  "temperature": "{Icon} Temperature {value}°C",
  "battery": "{Icon}",
  "last_seen": "formatLastSeen(value)", // 22 minutes ago
  "linkquality": "{Icon} {value} LQI"
};


function formatLastSeen(sensorLastSeen)
{
  // find diff between now and last_seen
    var lastSeen = moment(sensorLastSeen);
    var diff = moment().diff(lastSeen);
    var duration = moment.duration(diff);
    
    // TODO: handle days() > 0
    if (duration.hours() > 0){
        return duration.hours() + " hours ago";
    }
    if (duration.minutes() > 0){
        return duration.minutes() + " minutes ago";
    }
    if (duration.seconds() > 5){
        return duration.seconds() + " seconds ago";
    }

    return "just now";
}


function getUnit(sensor) {
  if (quickViewSensorReadings[sensor] == undefined) {
    return "";
  }
  return quickViewSensorReadings[sensor];
}

function format(sensor, value){
  return sensor + ": "+ value + " " + getUnit(sensor)
}

function getIcon(reading) {
}

function isWhitelisted(sensor) {

  return quickViewSensorReadings[sensor] != undefined;
}

</script>

<template>
  <ol>{{ device.name }}</ol>
  <ol v-for="(value, sensor) in device.payload">
    <ol v-if="isWhitelisted(sensor)">
        {{ format(sensor, value) }} 
    </ol>
  </ol>
</template>