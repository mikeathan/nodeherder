<script setup lang="ts">
  import { computed } from 'vue';
  import type { IntegrationResource } from '../types';
  import WeatherResults from '@/components/integrations/WeatherResults.vue';
  const props = defineProps<{ resource: IntegrationResource }>();
  const forecast = computed(() => (props.resource.contract_version === 'weather_forecast/v1' ? props.resource : null));
</script>
<template><WeatherResults v-if="forecast" :forecast="forecast" /></template>
