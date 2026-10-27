<script setup lang="ts">
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import { computed } from 'vue';
  import { useRoute } from 'vue-router';
  import { integrationReturnTarget } from '@/domain/integration-navigation';

  defineProps<{ title: string; subtitle?: string }>();
  const route = useRoute();
  const back = computed(() => integrationReturnTarget(route.query));
</script>

<template>
  <UiPageHeader class="integration-page-heading" :title="title" :back="back.to" :back-label="back.label">
    <!-- Reserve one subtitle line before metadata arrives; it must not move the section links. -->
    <template #subtitle>
      <span v-if="subtitle">{{ subtitle }}</span>
      <span v-else aria-hidden="true">&nbsp;</span>
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
