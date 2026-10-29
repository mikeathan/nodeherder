<script setup lang="ts">
  import WasteBinIcon from '@/components/integrations/WasteBinIcon.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  defineProps<{ kind: string; colour: string | null }>();
</script>
<template><WasteBinIcon v-if="kind === 'waste_service'" :colour="colour" /><UiIcon v-else name="schedule" /></template>
