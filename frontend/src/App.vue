<script setup>
import { ref } from 'vue'
import { useStore } from 'vuex'
import { computed } from "vue";
import SensorData from './components/Dashboard.vue'



const socketUri = "ws://localhost:3000/ws" //document.location.host
console.log("sockeruri:" + socketUri)
const socket = new WebSocket(socketUri)
const title = ref('Node-Herder Page')
const store = useStore()
const devices = computed(() => store.getters.devices)

socket.onmessage = (event) => {

  const obj = JSON.parse(event.data);
  if (event.data == undefined) {
    return
  }
  console.log("message:" + event.data)
  console.log("message:" + obj.name + " " + obj.payload)
  store.commit("deviceUpdated", obj);
 
}

socket.onopen = function (event) {
  console.log("Open: " + event);
}
socket.onclose = function (event) {
  console.log("Close");
}
socket.onerror = function (event) {

  console.log("Error: " + event.data);
}

const deviceData = ref([
  {
    "name": "TH1",
    "payload":
    {
      "battery": 100,
      "humidity": 69.8,
      "last_seen": "2023-05-06T19:13:37+01:00",
      "linkquality": 29,
      "temperature": 20,
      "voltage": 3000
    }
  }, {
    "name": "TH2",
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
  <h1>{{ title }}</h1>
  <ol>
    <SensorData v-for="(item) in devices" :device="item" :key="item.name"></SensorData>
  </ol>
</template>

