<script setup lang="ts">
  import { computed } from 'vue';
  import { formatSmartDate, parseTimestamp } from '@/utils/date.utils';

  const props = defineProps({
    from: {
      type: [String, Number],
      required: false,
    },
    to: {
      type: [String, Number],
      required: false,
    },
  });

  const formattedRange = computed(() => {
    if (!props.from || !props.to) return '';
    const fromTs = parseTimestamp(props.from, 0);
    const toTs = parseTimestamp(props.to, 0);
    if (!fromTs || !toTs) return '';
    return `${formatSmartDate(fromTs)} - ${formatSmartDate(toTs)}`;
  });
</script>

<template>
  <div v-if="formattedRange" class="date-range-display text-caption text-medium-emphasis mb-2">
    <i class="pi pi-calendar mr-1" style="font-size: 0.875rem"></i>
    {{ formattedRange }}
  </div>
</template>

<style scoped>
  .date-range-display {
    display: flex;
    align-items: center;
  }
</style>
