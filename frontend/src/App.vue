<script setup>
import { ref } from "vue";
import Dashboard from "./components/device-dashboard/DeviceDashboard.vue";
const title = ref("Node-Herder");
</script>
<style></style>

<template>
  <main className="content p-0 p-sm-3">
    <div class="container-fluid p-0 h-100">
      <a href="#/">{{ title }}</a>
      <RouterView />
    </div>
  </main>
</template>
