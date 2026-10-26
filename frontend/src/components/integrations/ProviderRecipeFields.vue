<script setup lang="ts">
  import { useEditableModel } from '@/composables/integrations/useEditableModel';
  import { humanLabel } from '@/domain/integrations';
  import type { ProviderDefinition } from '@/types/integrations.type';
  const props = defineProps<{
    modelValue: Pick<ProviderDefinition, 'setup' | 'refresh'>;
    requestNames: string[];
    inputs: ProviderDefinition['inputs'];
  }>();
  const emit = defineEmits<{ 'update:modelValue': [value: Pick<ProviderDefinition, 'setup' | 'refresh'>] }>();
  const fields = useEditableModel(
    () => props.modelValue,
    (value) => emit('update:modelValue', value)
  );
</script>
<template>
  <div class="connection-form">
    <details>
      <summary>Connect the setup and refresh steps</summary>
      <div class="connection-form">
        <label
          v-for="(step, key) in {
            search: fields.setup.search,
            select: fields.setup.select,
            status: fields.setup.status,
            refresh: fields.refresh,
          }"
          :key="key"
          class="connection-field"
          ><span>{{
            key === 'search'
              ? 'Find addresses'
              : key === 'select'
                ? 'Retrieve selected address'
                : key === 'status'
                  ? 'Check lookup status'
                  : 'Read saved schedule'
          }}</span
          ><select v-model="step.request">
            <option v-for="id in requestNames" :key="id" :value="id">{{ humanLabel(id) }}</option>
          </select></label
        ><label class="connection-field"
          ><span>Request a fresh lookup when the saved schedule expires</span
          ><select v-model="fields.refresh.renewal_request">
            <option value="">No renewal request</option>
            <option v-for="id in requestNames" :key="id" :value="id">{{ humanLabel(id) }}</option>
          </select></label
        >
        <div class="connection-row">
          <label class="connection-field"
            ><span>Search input</span
            ><select v-model="fields.setup.search.input">
              <option v-for="(input, key) in inputs" :key="key" :value="key">{{ input.label }}</option>
            </select></label
          ><label class="connection-field"
            ><span>Address choice input</span
            ><select v-model="fields.setup.select.input">
              <option v-for="(input, key) in inputs" :key="key" :value="key">{{ input.label }}</option>
            </select></label
          >
        </div>
        <label v-for="(value, key) in fields.setup.search.options" :key="key" class="connection-field"
          ><span>Address list: {{ humanLabel(String(key)) }}</span
          ><input v-model="fields.setup.search.options[key]"
        /></label>
        <div v-if="fields.setup.search.expect" class="connection-row">
          <label class="connection-field"
            ><span>Required input field location</span><input v-model="fields.setup.search.expect.pointer" /></label
          ><label class="connection-field"
            ><span>Supported input value</span><input v-model="fields.setup.search.expect.value"
          /></label>
        </div>
        <label class="connection-field"
          ><span>Lookup status field location</span><input v-model="fields.setup.status.pointer" /></label
        ><label v-for="key in ['pending', 'done', 'failed'] as const" :key="key" class="connection-field"
          ><span>{{ humanLabel(key) }} status values — separated by commas</span
          ><input
            :value="fields.setup.status[key].join(', ')"
            @change="
              fields.setup.status[key] = ($event.target as HTMLInputElement).value
                .split(',')
                .map((s) => s.trim())
                .filter(Boolean)
            "
        /></label>
      </div>
    </details>
  </div>
</template>
