<script setup lang="ts">
  import { computed } from 'vue';

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

    const fromDate = new Date(props.from);
    const toDate = new Date(props.to);
    const now = new Date();

    return `${formatSmartDate(fromDate)} – ${formatSmartDate(toDate)}`;
  });

  function formatSmartDate(date: Date): string {
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    const dateOnly = new Date(date.getFullYear(), date.getMonth(), date.getDate());

    const timeStr = date.toLocaleTimeString('en-GB', {
      hour: '2-digit',
      minute: '2-digit',
    });

    if (dateOnly.getTime() === today.getTime()) {
      return `Today ${timeStr}`;
    }

    if (dateOnly.getTime() === yesterday.getTime()) {
      return `Yesterday ${timeStr}`;
    }

    // Same year
    if (date.getFullYear() === now.getFullYear()) {
      const dateStr = date.toLocaleDateString('en-GB', {
        day: '2-digit',
        month: 'short',
      });
      return `${dateStr} ${timeStr}`;
    }

    // Different year
    const dateStr = date.toLocaleDateString('en-GB', {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
    });
    return `${dateStr} ${timeStr}`;
  }
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
