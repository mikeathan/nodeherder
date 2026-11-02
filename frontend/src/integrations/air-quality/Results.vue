<script setup lang="ts">
  import { computed } from 'vue';
  import type { IntegrationResource } from '../types';
  import AirQualityResults from '@/components/integrations/air-quality/AirQualityResults.vue';

  const props = defineProps<{ resource: IntegrationResource }>();
  const forecast = computed(() => (props.resource.contract_version === 'air_quality/v1' ? props.resource : null));
</script>

<template>
  <AirQualityResults v-if="forecast" :forecast="forecast" />
  <p v-else class="aq-muted">No saved air-quality data.</p>
</template>
