<script setup>


const props = defineProps({
  device: Object
})

const expectedSensorTypes = {
  "temperature": "Temperature!!!",
  "pressure": "Pressure!!!",
  "humidity": "Humidity!!!",
  "battery": "battery!!!!",
  "last_seen": "last_seen!!!!",
  "linkquality": "linkquality!!!!",
  "voltage": "voltage!!!!"
};
function renderDevice(device){

  for (var value of device) 
  {
    console.log("debug "  +value);
  }
}
</script>

<template>
  <ol>{{ device }}</ol>
</template>