<script setup lang="ts">
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import { RouteName } from '@/types/router';
  import './connections.css';
</script>
<template>
  <div class="connections">
    <UiPageHeader title="Integration page not found" />
    <p>This integration page is unavailable.</p>
    <UiButton :to="{ name: RouteName.Integrations }">Back to Integrations</UiButton>
  </div>
</template>
