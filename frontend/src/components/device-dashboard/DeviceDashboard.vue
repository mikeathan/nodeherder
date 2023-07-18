<script setup>
import { useStore } from "vuex";
import { computed } from "vue";
import Device from "./DeviceCard.vue";

const store = useStore();
const devices = computed(() => store.getters.devices);
</script>

<template>
  <div className="container-fluid p-0 h-100">
    <div className="row">
      <Device v-for="item in devices" :device="item" :key="item.name"></Device>
    </div>


  </div>
</template>
