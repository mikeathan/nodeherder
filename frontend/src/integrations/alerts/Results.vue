<script setup lang="ts">
  import { computed } from 'vue';
  import type { IntegrationResource } from '../types';
  import AlertsResults from '@/components/integrations/alerts/AlertsResults.vue';

  const props = defineProps<{ resource: IntegrationResource }>();
  const alerts = computed(() => (props.resource.contract_version === 'weather_alerts/v1' ? props.resource : null));
</script>

<template>
  <AlertsResults v-if="alerts" :alerts="alerts" />
  <p v-else class="alert-muted">No saved weather alerts.</p>
</template>
