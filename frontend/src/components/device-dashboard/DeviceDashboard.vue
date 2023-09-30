<script setup>
import { useStore } from "vuex";
import { computed } from "vue";
import DeviceCard from "./DeviceCard.vue";

const store = useStore();
const devices = computed(() => store.getters.devices);




</script>

<template>
  <div className="row">
    <DeviceCard v-for="item in devices" :device="item" :key="item.name"></DeviceCard>
  </div>
</template>
