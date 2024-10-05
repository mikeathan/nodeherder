<script setup>

import { onBeforeMount, onMounted, onUnmounted, ref } from "vue";
import { store } from "./store/index";
import Status from "./components/controls/Status.vue";

const title = ref("Node-Herder");
const intervalId = 0;
const EXPIRATION_TIMEOUT = 10 * 60 * 1000 // 10  min
onBeforeMount(() => {
  store.dispatch('ws/connect')
});

onMounted(() => {
  intervalId = setInterval(() => {
    console.log('Start ExpirationCheck');

    store.dispatch('deleteExpiredMessages');
  }, EXPIRATION_TIMEOUT);

  console.log('ExpirationCheck id:', intervalId);

});

onUnmounted(() => {
  console.log('Clear ExpirationChec id:', intervalId);

  clearInterval(intervalId);
})
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
      <RouterLink to="/viewer">Automations </RouterLink> |
      <RouterLink to="/settings">Settings </RouterLink>

      <RouterView />
    </div>
  </main>
</template>
