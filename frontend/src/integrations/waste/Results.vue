<script setup lang="ts">
  import { computed } from 'vue';
  import type { IntegrationResource } from '../types';
  import WasteScheduleResults from '@/components/integrations/WasteScheduleResults.vue';
  const props = defineProps<{ resource: IntegrationResource }>();
  const schedule = computed(() => (props.resource.contract_version === 'waste_schedule/v1' ? props.resource : null));
</script>
<template><WasteScheduleResults v-if="schedule" :resources="[schedule]" /></template>
