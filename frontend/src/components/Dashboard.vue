<script setup>
import { useStore } from "vuex";
import { computed } from "vue";
import Device from "./Device.vue";

const store = useStore();
const devices = computed(() => store.getters.devices);
</script>

<template>
  <ol>
    <Device v-for="item in devices" :device="item" :key="item.name"></Device>
  </ol>
</template>
