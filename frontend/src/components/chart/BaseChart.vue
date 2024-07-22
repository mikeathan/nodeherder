<script setup lang="ts">
import { ref, onBeforeMount } from 'vue';
import type { PropType, Ref } from 'vue';
import VueApexCharts from 'vue3-apexcharts';

const props = defineProps({
  data: {
    type: Object as PropType<any>,
    default: null,
  },
  options: {
    type: Object,
    default: null,
  },
});

onBeforeMount(() => {});
</script>

<template>
  <div class="chart-container">
    <VueApexCharts
      :options="props.options"
      :series="props.data">
    </VueApexCharts>
  </div>
</template>
