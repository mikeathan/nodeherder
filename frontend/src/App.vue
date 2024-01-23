<script setup>

import { onBeforeMount, ref } from "vue";
import { store } from "./store/index";
import Status from "./components/controls/Status.vue";

const title = ref("Node-Herder");

onBeforeMount(() => {
  store.dispatch('ws/connect')
});

</script>
<style></style>

<template>
  <main className="content p-0 p-sm-3">
    <div class="container-fluid p-0 h-100">
      <span class="me-1">
        <Status></Status>
      </span>
      <Notifications position="top right" />
      <RouterLink to="/">{{ title }}</RouterLink> |
      <RouterLink to="/viewer">Automations </RouterLink>
      <RouterView />
    </div>
  </main>
</template>
