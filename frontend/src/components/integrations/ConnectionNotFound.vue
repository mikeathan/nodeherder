<script setup lang="ts">
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import { RouteName } from '@/types/router';
  import './connections.css';
</script>
<template>
  <div class="connections">
    <UiPageHeader title="Connection page not found" />
    <p>This connection page is unavailable.</p>
    <UiButton :to="{ name: RouteName.Connections }">Back to Connections</UiButton>
  </div>
</template>
