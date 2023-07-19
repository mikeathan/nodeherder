<script setup>
import { useStore } from "vuex";
import { computed } from "vue";
import Device from "./DeviceCard.vue";

const store = useStore();
const devices = computed(() => store.getters.devices);
</script>

<template>
  <div className="row">
    <Device v-for="item in devices" :device="item" :key="item.name"></Device>
  </div>
</template>
