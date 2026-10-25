<script setup lang="ts">
  import UiButton from '@/components/ui/UiButton.vue';
  import { useEditableModel } from '@/composables/integrations/useEditableModel';
  import type { ProviderDefinition } from '@/types/integrations.type';
  const props = defineProps<{ modelValue: Pick<ProviderDefinition, 'policy' | 'errors'> }>();
  const emit = defineEmits<{ 'update:modelValue': [value: Pick<ProviderDefinition, 'policy' | 'errors'>] }>();
  const fields = useEditableModel(
    () => props.modelValue,
    (value) => emit('update:modelValue', value)
  );
</script>
<template>
  <div class="connection-form">
    <details>
      <summary>Request limits and errors</summary>
      <div class="connection-form">
        <label class="connection-field"
          ><span>Minimum seconds between status checks</span
          ><input v-model.number="fields.policy.poll_min_seconds" type="number" min="5" /></label
        ><label class="connection-field"
          ><span>Minimum seconds between new lookups</span
          ><input v-model.number="fields.policy.renewal_min_seconds" type="number" min="86400" /></label
        ><label class="connection-field"
          ><span>Provider's documented submission-key lifetime in seconds</span
          ><input v-model.number="fields.policy.idempotency_window_seconds" type="number" min="0" /></label
        ><label class="connection-field"
          ><span>Provider error code location</span><input v-model="fields.errors.code"
        /></label>
        <div v-for="(rule, i) in fields.errors.rules" :key="i" class="connection-form">
          <div class="connection-row">
            <label class="connection-field"
              ><span>HTTP error status</span
              ><input v-model.number="rule.status" type="number" min="400" max="599" /></label
            ><label class="connection-field"
              ><span>Exact provider problem code</span><input v-model="rule.code"
            /></label>
          </div>
          <label class="connection-field"
            ><span>Recovery action</span
            ><select v-model="rule.action">
              <option value="renewal">Request a fresh lookup</option>
              <option value="restart_setup">Ask to repeat setup</option>
              <option value="reconnect">Ask to replace the credential</option>
              <option value="unsupported">Explain that this setup is unsupported</option>
            </select></label
          >
          <div><UiButton variant="ghost" @click="fields.errors.rules.splice(i, 1)">Remove error rule</UiButton></div>
        </div>
        <div>
          <UiButton @click="fields.errors.rules.push({ status: 404, code: '', action: 'restart_setup' })"
            >Add error rule</UiButton
          >
        </div>
      </div>
    </details>
  </div>
</template>
