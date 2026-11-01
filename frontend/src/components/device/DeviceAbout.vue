<script setup lang="ts">
  /* About card for one device; row content comes from domain/device-page.ts. */
  import { computed } from 'vue';
  import type { Device } from '@/types/device';
  import { deviceAboutRows } from '@/domain/devices/page';
  import UiCard from '@/components/ui/UiCard.vue';
  import RelativeTime from '@/components/ui/RelativeTime.vue';

  const props = defineProps<{ device: Device }>();
  const rows = computed(() => deviceAboutRows(props.device));
</script>

<template>
  <UiCard title="About" icon="info" flush>
    <dl class="nh-kv">
      <template v-for="row in rows" :key="row.key">
        <dt>{{ row.label }}</dt>
        <dd>
          <RelativeTime v-if="row.relative" :value="device.last_seen" />
          <template v-else-if="row.code"><code>{{ row.value }}</code></template>
          <template v-else>{{ row.value }}<span v-if="row.note" class="nh-muted">{{ row.note }}</span></template>
        </dd>
      </template>
    </dl>
  </UiCard>
</template>
