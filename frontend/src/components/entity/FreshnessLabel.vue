<script setup lang="ts">
  /*
   * Reporting-age label for a device. Presentation only: freshness is derived from
   * `last_seen` and never changes availability or command gating.
   */
  import { computed } from 'vue';
  import type { Device } from '@/types/device';
  import { freshnessPolicy, reportFreshness } from '@/domain/devices/freshness';
  import { useNow } from '@/composables/useClock';
  import RelativeTime from '@/components/ui/RelativeTime.vue';

  const props = withDefaults(
    defineProps<{ device: Device; mode?: 'updated' | 'seen'; compact?: boolean; short?: boolean }>(),
    { mode: 'updated', compact: false, short: false }
  );

  const now = useNow();
  const freshness = computed(() => reportFreshness(props.device.last_seen, now.value, freshnessPolicy(props.device)));

  /** Screen-reader and visible copy for one freshness band. */
  const prefix = computed(() => {
    if (props.short) return '';
    if (props.mode === 'seen') return 'Last seen';
    return freshness.value === 'stale' ? 'Last known' : 'Updated';
  });
  const suffix = computed(() => (freshness.value === 'quiet' ? ' · not reporting' : freshness.value === 'stale' ? ' · stale' : ''));
</script>

<template>
  <span
    v-if="freshness !== 'unknown'"
    class="nh-freshness"
    :class="`is-${freshness}`"
    :data-freshness="freshness">
    <RelativeTime v-if="compact" :value="device.last_seen" />
    <template v-else><template v-if="prefix">{{ prefix }} </template><RelativeTime :value="device.last_seen" />{{ suffix }}</template>
  </span>
  <span v-else-if="!compact" class="nh-freshness is-unknown">No report yet</span>
</template>

<style scoped>
  .nh-freshness {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    min-width: 0;
    color: var(--nh-text-muted);
    font-size: 0.7rem;
    line-height: 1.35;
  }
  .nh-freshness.is-quiet {
    color: var(--nh-warn);
  }
  .nh-freshness.is-stale {
    color: var(--nh-danger);
  }
  .nh-freshness.is-unknown {
    color: var(--nh-text-muted);
  }
</style>
