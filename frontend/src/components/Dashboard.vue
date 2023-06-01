<script setup>


const props = defineProps({
  device: Object
})

const excpectedUnits = {
  "temperature": "°C",
  "pressure": "hPa",
  "humidity": "%",
  "battery": "%",
  "last_seen": "last_seen!!!!",
  "linkquality": "lqi",
  "voltage": "mV"
};
function getUnit(reading) {

  if (excpectedUnits["key"] == undefined) {
    return "";
  }
  return excpectedUnits[reading];
  // check from list of units whats the unit for reading
}

function getIcon(reading) {
  // check from list of icons whats the icon for reading

}

function isWhitelisted(reading) {

  // check from list of unitsif reading is whitelisted for rendering
  return true;
}

</script>

<template>
  <ol>{{ device.name }}</ol>

  <ol v-for="(value, key) in device.payload">

    {{ key }}: {{ value }}
  </ol>
</template>