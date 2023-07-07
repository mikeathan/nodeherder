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
    <nav class="navbar navbar-expand-md navbar-light">
      <div class="container-fluid">
        <a href="#/">{{ title }}</a>
      </div>
    </nav>
    <Dashboard />
  </div>
</template>
