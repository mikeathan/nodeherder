<script setup lang="ts">
import { computed } from 'vue';
import { formatSmartDate, parseTimestamp } from '@/utils/date.utils';

const props = defineProps({
  from: [String, Number],
  to: [String, Number],
});

const formattedRange = computed(() => {
  if (!props.from || !props.to) return '';
  const fromTs = parseTimestamp(props.from, 0);
  const toTs = parseTimestamp(props.to, 0);
  if (!fromTs || !toTs) return '';
  return `${formatSmartDate(fromTs)} – ${formatSmartDate(toTs)}`;
});
</script>

<template>
  <div v-if="formattedRange" class="range-row">
    <i class="pi pi-calendar icon"></i>
    <span class="text">{{ formattedRange }}</span>
  </div>
</template>

<style scoped>
.range-row {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 0;
  font-size: 0.85rem;
  opacity: 0.75;
}

/* prevents pushing content down */
.range-row:not(:last-child) {
  margin-bottom: 0;
}

.icon {
  font-size: 0.85rem;
}

.text {
  white-space: nowrap;
}
</style>
