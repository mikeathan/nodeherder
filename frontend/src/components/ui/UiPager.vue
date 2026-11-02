<script setup lang="ts">
  import { computed } from 'vue';
  import UiButton from './UiButton.vue';

  /** Shared paged-result controls; labels and status copy stay owned by the caller. */
  const props = withDefaults(
    defineProps<{
      page: number;
      pageCount: number;
      label?: string;
      status?: string;
      firstLabel?: string;
      previousLabel?: string;
      nextLabel?: string;
      lastLabel?: string;
      showFirst?: boolean;
      showLast?: boolean;
      disabled?: boolean;
      announceStatus?: boolean;
    }>(),
    {
      label: 'Pages',
      status: '',
      firstLabel: 'First',
      previousLabel: 'Previous',
      nextLabel: 'Next',
      lastLabel: 'Last',
      showFirst: true,
      showLast: true,
      disabled: false,
      announceStatus: false,
    }
  );
  const emit = defineEmits<{ 'update:page': [page: number] }>();
  const count = computed(() => Math.max(1, props.pageCount));
  const current = computed(() => Math.min(Math.max(1, props.page), count.value));
  const canPrevious = computed(() => !props.disabled && current.value > 1);
  const canNext = computed(() => !props.disabled && current.value < count.value);
  const columns = computed(() => 2 + Number(props.showFirst) + Number(props.showLast));
  function go(page: number) {
    const next = Math.min(Math.max(1, page), count.value);
    if (next !== current.value) emit('update:page', next);
  }
</script>

<template>
  <nav
    v-if="pageCount > 1"
    class="ui-pager"
    :style="{ '--pager-columns': columns }"
    :aria-label="label">
    <div class="ui-pager-actions">
      <UiButton v-if="showFirst" :disabled="!canPrevious" @click="go(1)">{{ firstLabel }}</UiButton>
      <UiButton :disabled="!canPrevious" @click="go(current - 1)">{{ previousLabel }}</UiButton>
    </div>
    <span
      v-if="status"
      class="ui-pager-status"
      :aria-live="announceStatus ? 'polite' : undefined">{{ status }}</span
    >
    <div class="ui-pager-actions">
      <UiButton :disabled="!canNext" @click="go(current + 1)">{{ nextLabel }}</UiButton>
      <UiButton v-if="showLast" :disabled="!canNext" @click="go(count)">{{ lastLabel }}</UiButton>
    </div>
  </nav>
</template>

<style scoped>
  .ui-pager {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) auto;
    align-items: center;
    gap: 0.75rem;
  }
  .ui-pager-status {
    color: var(--nh-text-muted);
    font-size: 0.82rem;
    text-align: center;
  }
  .ui-pager-actions {
    display: flex;
    gap: 0.5rem;
  }
  .ui-pager-actions:last-child {
    justify-content: flex-end;
  }
  @media (max-width: 640px) {
    .ui-pager {
      grid-template-columns: repeat(var(--pager-columns, 4), minmax(0, 1fr));
      gap: 0.6rem 0.4rem;
    }
    .ui-pager-status {
      grid-column: 1 / -1;
      grid-row: 1;
    }
    .ui-pager-actions {
      display: contents;
    }
    .ui-pager :deep(.nh-btn) {
      grid-row: 2;
      width: 100%;
      min-width: 0;
      padding: 0 0.25rem;
      font-size: 0.78rem;
    }
  }
</style>
