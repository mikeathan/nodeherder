<script setup>
import { ref } from "vue";
import Dashboard from "./components/device-dashboard/DeviceDashboard.vue";
const title = ref("Node-Herder");
</script>
<style>
div {
  padding-left: 10px;
}
</style>

<template>
  <div class="main">
    <div class="container-fluid">
      <a href="#/">{{ title }}</a>
    </div>
    <RouterView />
  </div>
</template>
