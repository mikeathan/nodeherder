<script setup>

import { onBeforeMount, onMounted, onUnmounted, ref } from "vue";
import { store } from "./store/index";
import Status from "./components/controls/Status.vue";
import Notifications from "./components/hub/alerts/Notifications.vue";

const title = ref("Node-Herder");

onBeforeMount(() => {
  store.dispatch('ws/connect')
});


</script>
<style></style>

<template>
  <v-app>

    <v-main>
      <v-app-bar dark>
        <v-toolbar-title>

            <span class="me-1">
              <Status></Status>
            </span>
            <Notifications />
            <RouterLink to="/">{{ title }}</RouterLink> |
            <RouterLink to="/viewer">Automations </RouterLink> |
            <RouterLink to="/consoleviewer">Console</RouterLink> |
            <RouterLink to="/settings">Settings </RouterLink>
        </v-toolbar-title>

      </v-app-bar>

      <RouterView />
    </v-main>

    <!-- <v-main className="content p-0 p-sm-3">
      
      <div class="container-fluid p-0 h-100">
        <span class="me-1">
          <Status></Status>
        </span>
        <Notifications />
        <RouterLink to="/">{{ title }}</RouterLink> |
        <RouterLink to="/viewer">Automations </RouterLink> |
        <RouterLink to="/consoleviewer">Console</RouterLink> |
        <RouterLink to="/settings">Settings </RouterLink>

        <RouterView />
      </div>
    </v-main> -->
  </v-app>
</template>
