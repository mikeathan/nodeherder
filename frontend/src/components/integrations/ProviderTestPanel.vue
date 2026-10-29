<script setup lang="ts">
  import UiButton from '@/components/ui/UiButton.vue';
  import type { ConnectionInstance, ConnectionSetup, ProviderTestMode } from '@/types/integrations.type';
  defineProps<{ instances: ConnectionInstance[]; preview: ConnectionSetup | null; busy: boolean }>();
  const instance = defineModel<string>('instance', { required: true });
  const sample = defineModel<string>('sample', { required: true });
  const emit = defineEmits<{ test: [mode: ProviderTestMode] }>();
</script>
<template>
  <details>
    <summary>Test without changing saved results</summary>
    <div class="connection-form">
      <label class="connection-field"
        ><span>Use inputs from this integration</span
        ><select v-model="instance" :disabled="busy">
          <option value="">Choose a saved integration</option>
          <option v-for="item in instances" :key="item.id" :value="item.id">{{ item.name }}</option>
        </select></label
      >
      <p class="connection-muted">
        Offline test checks a supplied schedule JSON sample. Live test sends only the saved-schedule GET request with
        private inputs; it does not start a new lookup or replace saved collection dates.
      </p>
      <label class="connection-field"
        ><span>Schedule JSON sample for an offline test</span>
        <textarea v-model="sample" :disabled="busy" spellcheck="false" rows="6" maxlength="262144" />
      </label>
      <div class="connection-actions">
        <UiButton :disabled="busy || !instance || !sample" @click="emit('test', 'offline')">Test sample</UiButton>
        <UiButton :disabled="busy || !instance" @click="emit('test', 'live')">Test saved-schedule request</UiButton>
      </div>
      <p v-if="preview" role="status">
        {{
          preview.last_error?.message ??
          (preview.preview ? 'Test complete. Saved results were not changed.' : 'Test in progress…')
        }}
      </p>
      <ul v-if="preview?.preview" class="connection-list">
        <li v-for="service in ('services' in preview.preview ? preview.preview.services : [])" :key="service.service_id">
          {{ service.name }}: {{ service.dates.join(', ') || 'dates unavailable' }}
        </li>
      </ul>
    </div>
  </details>
</template>
