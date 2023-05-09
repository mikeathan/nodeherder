<script setup>
import { ref } from 'vue'
import SensorData from './components/SensorData.vue'

const message = ref('Node-Herder Page')
const deviceData = ref([
{
    "device_name": "TH1",
    "payload":
    {
        "battery": 100,
        "humidity": 69.8,
        "last_seen": "2023-05-06T19:13:37+01:00",
        "linkquality": 29,
        "temperature": 20,
        "voltage": 3000
    }
},{
    "device_name": "TH2",
    "payload":
    {
    "battery": 100,
    "humidity": 61.8,
    "last_seen": "2023-05-09T17:07:22+01:00",
    "linkquality": 32,
    "temperature": 22.2,
    "voltage": 3000
    }
}
]);

</script>

<template>
  <h1>{{ message }}</h1>
  <ol> 
  <SensorData 
  v-for="(item) in deviceData"
      :device="item"
      :key="item.device_name"
    ></SensorData>
  </ol>
</template>

