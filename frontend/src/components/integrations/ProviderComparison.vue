<script setup lang="ts">
  import { computed } from 'vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import { providerChanges } from '@/domain/integrations';
  import type { ConnectionDefinition } from '@/types/integrations.type';
  const props = defineProps<{ provider: ConnectionDefinition; busy: boolean }>();
  const emit = defineEmits<{ useShipped: []; keepVersion: [] }>();
  const comparison = computed(() => providerChanges(props.provider.yaml, props.provider.shipped_yaml ?? ''));
</script>
<template>
  <details>
    <summary>
      {{ provider.update_available ? 'Updated supplied example available' : 'Compare with the supplied example' }}
    </summary>
    <p class="connection-muted">
      App updates keep your host file. Review the differences before using the supplied example; it replaces only the
      draft until you Save changes.
    </p>
    <pre class="connection-comparison">{{ comparison }}</pre>
    <div class="connection-actions">
      <UiButton :disabled="busy" @click="emit('useShipped')">Use supplied example in editor</UiButton>
      <UiButton v-if="provider.update_available" :disabled="busy" @click="emit('keepVersion')"
        >Keep my version</UiButton
      >
    </div>
  </details>
</template>
