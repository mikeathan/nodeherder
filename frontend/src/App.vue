<script setup>
import { ref } from 'vue'
import SensorData from './components/SensorData.vue'

const message = ref('Node-Herder Page')
const deviceData = ref([{
    "battery": 100,
    "humidity": 69.8,
    "last_seen": "2023-05-06T19:13:37+01:00",
    "linkquality": 29,
    "temperature": 20,
    "voltage": 3000
}]);
const groceryList = ref([
  { id: 0, note: 'Vegetables' },
  { id: 1, note: 'Cheese' },
  { id: 2, note: 'Whatever else humans are supposed to eat' }
])
</script>

<template>
  <h1>{{ message }}</h1>
  <div>{{ deviceData }}</div>
  <ol>
  <SensorData
  v-for="item in groceryList"
      :device="item"
      :key="item.id"
    ></SensorData>
  </ol>
</template>

