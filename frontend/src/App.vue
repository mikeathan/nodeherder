<script setup>
import { ref } from 'vue'
import SensorData from './components/Dashboard.vue'

const socket = new WebSocket("ws://" + document.location.host + "/ws")
const message = ref('Node-Herder Page')
socket.onmessage = (event) => {
  const message = JSON.parse(event.data);
  print("message:" + message)
}
socket.onopen = function(event) {
  print("Open");
}
socket.onclose = function(event) {
  print("Close");
  socket = null;
}
socket.onerror = function(event) {
  print("Error: " + event.data);
}

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

