<script setup>
import { stringLiteral } from '@babel/types';



const props = defineProps({
  device: Object
})


const deviceReadingsWhitelist = {
  "voltage": "mV",
  "state":" .", // this is just for testing - remove
  "battery": "%",       // icon and caption
  "last_seen": "(WIP)", // 22 minutes ago
  "linkquality": "LQI", // icon and caption
}

const sensorReadingsWhitelist= {
  "temperature": "${icon}Temperature ${value}°C",
  "pressure": "${icon}Pressure ${value}hPa",
  "humidity": "${icon}Humidity ${value}%",  
};

const sensorUnits= {
  "temperature": "°C",
  "pressure": "$hPa",
  "humidity": "%",  
  "voltage": "mV",
  "linkquality": "LQI"
};

// https://github.com/nurikk/zigbee2mqtt-frontend/blob/dev/src/components/dashboard-page/index.tsx

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
  if (sensorUnits[sensor] == undefined) {
    return "";
  }
  return sensorUnits[sensor];
}

function format(sensor, value){
  
  // icon type unit
  var sensorName =  sensor.charAt(0).toUpperCase() + sensor.slice(1);
  return sensorName + " "+ value + getUnit(sensor);
}

function getIcon(reading) {
  //https://github.com/nurikk/zigbee2mqtt-frontend/blob/dev/src/components/dashboard-page/DashboardFeatureWrapper.tsx
   return "fa-thermometer-full";
}

function isSensorReading(reading) {
  return sensorReadingsWhitelist[reading] != undefined;
}

function isDeviceReading(reading) {
  return deviceReadingsWhitelist[reading] != undefined;
}

</script>

<template>
  <ol>{{ device.name }}</ol>
  <ol v-for="(value, sensor) in device.payload">
    <ol v-if="isSensorReading(sensor)">

      <!-- <div>
          <i  v-bind:class="{{ getIcon(sensor)  }}" >
          </i>
        </div> -->
      {{ getIcon(sensor) }} {{ format(sensor, value) }} 
    </ol>
  </ol>
</template>