<script setup lang="ts">
  import { useEditableModel } from '@/composables/integrations/useEditableModel';
  import { humanLabel } from '@/domain/integrations';
  import type { ProviderDefinition } from '@/types/integrations.type';
  const props = defineProps<{ modelValue: Pick<ProviderDefinition, 'name' | 'base_url' | 'auth' | 'inputs'> }>();
  const emit = defineEmits<{
    'update:modelValue': [value: Pick<ProviderDefinition, 'name' | 'base_url' | 'auth' | 'inputs'>];
  }>();
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
    <fieldset class="connection-form">
      <legend>Provider</legend>
      <RouterLink :to="{ name: RouteName.ConnectionHelp, params: route.params, query: route.query, hash: '#repair' }"
        >Help changing a URL</RouterLink
      ><label class="connection-field"
        ><span>Display name</span><input id="name" v-model="fields.name" maxlength="120" /></label
      ><label class="connection-field"
        ><span>Provider URL</span><input id="base_url" v-model="fields.base_url" type="url"
      /></label>
      <div class="connection-row">
        <label class="connection-field"
          ><span>Authentication</span
          ><select v-model="fields.auth.mode">
            <option value="none">No token</option>
            <option value="optional_bearer">Optional API token</option>
            <option value="bearer">Required API token</option>
          </select></label
        ><label v-if="fields.auth.mode !== 'none'" class="connection-field"
          ><span>Credential slot name</span><input v-model="fields.auth.secret_slot"
        /></label>
      </div>
      <p class="connection-muted">
        The provider ID and supported output type stay the same when repairing this file. Use Copy provider to create a
        separate definition.
      </p>
    </fieldset>
    <details>
      <summary>Setup inputs</summary>
      <div class="connection-form">
        <div v-for="(input, key) in fields.inputs" :key="key" class="connection-row">
          <label class="connection-field"
            ><span>{{ humanLabel(String(key)) }} label</span><input v-model="input.label" /></label
          ><label class="connection-field"
            ><span>Input type</span
            ><select v-model="input.type">
              <option value="postcode">Postcode</option>
              <option value="select">Address choice</option>
              <option value="text">Text</option>
            </select></label
          ><label><input v-model="input.required" type="checkbox" /> Required</label>
        </div>
      </div>
    </details>
  </div>
</template>
