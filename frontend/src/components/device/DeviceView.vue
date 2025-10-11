<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { store } from '@/store';
import DeviceCard from '@/components/dashboards/cards/DeviceCard.vue';
import type { Device } from '@/types/device';

const route = useRoute();
const id = computed(() => String(route.params.id));

const device = computed<Device | undefined>(() =>
  store.getters['hub/findDevice'](id.value) as Device | undefined
);
</script>

<template>
  <div class="device-view">
    <div class="card-wrap" v-if="device">
      <DeviceCard :device="device" />
    </div>
    <div v-else>Device not found.</div>
  </div>
</template>

<style scoped>
.device-view { padding: 1rem; }
.card-wrap {
  --max-width: 640px;
  inline-size: min(100%, var(--max-width));
  margin-inline: auto;
}
</style>