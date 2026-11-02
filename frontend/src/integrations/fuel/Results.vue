<script setup lang="ts">
  import { computed } from 'vue';
  import type { IntegrationResource } from '../types';
  import FuelResults from '@/components/integrations/fuel/FuelResults.vue';

  const props = defineProps<{ resource: IntegrationResource }>();
  const fuel = computed(() => (props.resource.contract_version === 'fuel_prices/v1' ? props.resource : null));
</script>

<template>
  <FuelResults v-if="fuel" :fuel="fuel" />
  <p v-else>No saved fuel prices.</p>
</template>
