<script setup>
import { ref } from "vue";
import Dashboard from "./components/device-dashboard/DeviceDashboard.vue";
const title = ref("Node-Herder Page");
</script>

<template>
  <h1>{{ title }}</h1>
  <Dashboard />
</template>
