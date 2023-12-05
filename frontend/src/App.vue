<script setup>
import { Alert } from "bootstrap";

import { onBeforeMount, ref } from "vue";
import { useStore } from "vuex";
import Status from "./components/output/Status.vue";
const title = ref("Node-Herder");
const store = useStore()

onBeforeMount(() => {
  store.dispatch('ws/connect')
});

</script>
<style></style>

<template>
  <main className="content p-0 p-sm-3">
    <div class="container-fluid p-0 h-100">
      <Status></Status>
      <notifications position="top right" />
      <RouterLink to="/">{{ title }}</RouterLink> |
      <RouterLink to="/viewer">Automations </RouterLink>
      <RouterView />
    </div>
  </main>
</template>
