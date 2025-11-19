<script setup lang="ts">
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
    width: {
      type: [String, Number],
      default: '100%',
    },
    height: {
      type: [String, Number],
      default: 260,
    },
  });
</script>

<template>
  <VueApexCharts :options="props.options" :series="props.data" :width="props.width" :height="props.height">
  </VueApexCharts>
</template>
