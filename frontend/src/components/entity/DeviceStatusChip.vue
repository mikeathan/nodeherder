<script setup lang="ts">
  import { computed } from 'vue';
  import { Device } from '@/types/device';
  import { statusOf } from '@/composables/useHub';
  import UiChip from '@/components/ui/UiChip.vue';

  const props = defineProps<{ device: Device }>();
  const status = computed(() => statusOf(props.device));
  const view = {
    online: { tone: 'ok', icon: 'check', text: 'Online' },
    offline: { tone: 'danger', icon: 'offline', text: 'Offline' },
    disabled: { tone: 'muted', icon: 'disabled', text: 'Disabled' },
    unknown: { tone: 'muted', icon: 'unknown', text: 'Unknown' },
  } as const;
</script>

<template>
  <UiChip :tone="view[status].tone" :icon="view[status].icon">{{ view[status].text }}</UiChip>
</template>
