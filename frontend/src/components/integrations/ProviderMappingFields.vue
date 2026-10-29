<script setup lang="ts">
  import { useEditableModel } from '@/composables/integrations/useEditableModel';
  import { humanLabel } from '@/domain/integrations';
  import type { ProviderDefinition } from '@/types/integrations.type';
  const props = defineProps<{ modelValue: Extract<ProviderDefinition['mapping'], { collections: unknown }> }>();
  const emit = defineEmits<{ 'update:modelValue': [value: Extract<ProviderDefinition['mapping'], { collections: unknown }>] }>();
  const fields = useEditableModel(
    () => props.modelValue,
    (value) => emit('update:modelValue', value)
  );
  import { useRoute } from 'vue-router';
  import { RouteName } from '@/types/router';
  const route = useRoute();
</script>
<template>
  <div class="connection-form">
    <details>
      <summary>Collection fields</summary>
      <RouterLink :to="{ name: RouteName.IntegrationHelp, params: route.params, query: route.query, hash: '#fields' }"
        >Help mapping collection fields</RouterLink
      >
      <div class="connection-form">
        <p class="connection-muted">
          Map the collection list once. The same fields are used for every returned bin; no separate configuration is
          needed for each colour.
        </p>
        <label class="connection-field"
          ><span>Collection list location</span
          ><input id="mapping.collections.items" v-model="fields.collections.items"
        /></label>
        <div class="connection-row">
          <label v-for="(value, key) in fields.collections.fields" :key="key" class="connection-field"
            ><span>{{ humanLabel(String(key)) }}</span
            ><input :id="`mapping.collections.fields.${key}`" v-model="fields.collections.fields[key]"
          /></label>
        </div>
        <h3>Schedule information</h3>
        <div class="connection-row">
          <label v-for="(value, key) in fields.metadata" :key="key" class="connection-field"
            ><span>{{ humanLabel(String(key)) }}</span
            ><input :id="`mapping.metadata.${key}`" v-model="fields.metadata[key]"
          /></label>
        </div>
      </div>
    </details>
  </div>
</template>
