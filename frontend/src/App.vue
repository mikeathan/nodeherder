<script setup>
import { ref } from 'vue'
import { useStore } from 'vuex'
import { computed } from "vue";
import SensorData from './components/Dashboard.vue'



const socketUri = "ws://localhost:3000/ws" 
//const socketUri = "ws://"+document.location.host+"/ws"
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


</script>

<template>
  <h1>{{ title }}</h1>
  <ol>
    <SensorData v-for="(item) in devices" :device="item" :key="item.name"></SensorData>
  </ol>
</template>

