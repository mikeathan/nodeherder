<script setup>


const props = defineProps({
  device: Object
})

const additionalSensorReadings = {
  "voltage": "mV"
}

const expectedSensorReadings= {
  "temperature": "°C",
  "pressure": "hPa",
  "humidity": "%",
  "battery": "%",      // icon and caption
  "last_seen": "(WIP)", // 22 minutes ago
  "linkquality": "lqi", // icon and caption
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
  if (expectedSensorReadings[sensor] == undefined) {
    return "";
  }
  return expectedSensorReadings[sensor];
}

function format(sensor, value){
  return sensor + ": "+ value + " " + getUnit(sensor)
}

function getIcon(reading) {
  // check from list of icons whats the icon for reading
}

function isWhitelisted(sensor) {

  return expectedSensorReadings[sensor] != undefined;
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