<script setup>
import { ref } from "vue";
import Dashboard from "./components/device-dashboard/DeviceDashboard.vue";
const title = ref("Node-Herder");
</script>
<style></style>

<template>
  <div className="main">
    <main className="content p-0 p-sm-3">
      <div class="container-fluid">
        <a href="#/">{{ title }}</a>
      </div>
    </main>
    <RouterView />
  </div>
</template>
