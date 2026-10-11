<script setup lang="ts">
  /* "5 min ago" that stays current; the exact time is in the tooltip. */
  import { computed } from 'vue';
  import { relativeTime } from '@/domain/time';
  import { useNow } from '@/composables/useClock';

  const props = defineProps<{ value: string | number | null | undefined }>();
  const now = useNow();
  const iso = computed(() => {
    const t = typeof props.value === 'number' ? props.value : props.value ? Date.parse(props.value) : NaN;
    return Number.isFinite(t) ? new Date(t).toISOString() : undefined;
  });
</script>

<template>
  <time :datetime="iso" :title="iso ? new Date(iso).toLocaleString() : undefined">{{ relativeTime(value, now) }}</time>
</template>
