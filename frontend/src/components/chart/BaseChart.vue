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
  });
</script>

<template>
  <VueApexCharts
    :options="props.options"
    :series="props.data">
  </VueApexCharts>
</template>
