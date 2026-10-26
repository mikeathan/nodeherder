<script setup lang="ts">
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import { RouteName } from '@/types/router';

  defineProps<{ title: string; subtitle?: string }>();
</script>

<template>
  <UiPageHeader
    class="integration-page-heading"
    :title="title"
    :back="{ name: RouteName.Integrations }"
    back-label="Integrations">
    <!-- Reserve one subtitle line before metadata arrives; it must not move the section links. -->
    <template #subtitle>
      <span v-if="subtitle">{{ subtitle }}</span><span v-else aria-hidden="true">&nbsp;</span>
    </template>
    <template v-if="$slots.actions" #actions><slot name="actions" /></template>
  </UiPageHeader>
</template>

<style scoped>
  .integration-page-heading :deep(.nh-page-titles) {
    /* Allow the compact Refresh action beside the title on mobile instead of a temporary extra row. */
    flex-basis: 0;
  }
</style>
