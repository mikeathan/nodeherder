<script setup lang="ts">
import { ref, onBeforeMount } from 'vue';
import type { PropType, Ref } from 'vue';

const props = defineProps({
  chartData: {
    type: Object as PropType<any>, // ChartData<'bar'
    default: null,
  },
  options: {
    type: Object,
    default: null,
  },
});

const chartOptions = ref({});

onBeforeMount(() => {});
</script>

<template></template>
