<script setup lang="ts">
  import { computed } from 'vue';
  import { useRoute } from 'vue-router';
  import UiButton from '@/components/ui/UiButton.vue';
  import ProviderIdentityFields from './ProviderIdentityFields.vue';
  import ProviderRequestFields from './ProviderRequestFields.vue';
  import ProviderRecipeFields from './ProviderRecipeFields.vue';
  import ProviderMappingFields from './ProviderMappingFields.vue';
  import ProviderPolicyFields from './ProviderPolicyFields.vue';
  import { humanLabel, valueChoices } from '@/domain/integrations';
  import { renameProviderRequest, renameProviderCapture, nextDefinitionKey } from '@/domain/provider-definition';
  import { RouteName } from '@/types/router';
  import type { ProviderDefinition, ProviderRequest } from '@/types/integrations.type';

  // This component coordinates controlled sections; sections never mutate this document's props.
  const props = defineProps<{ modelValue: ProviderDefinition }>();
  const emit = defineEmits<{ 'update:modelValue': [value: ProviderDefinition] }>();
  const route = useRoute();
  const choices = computed(() => valueChoices(props.modelValue));
  const requestNames = computed(() => Object.keys(props.modelValue.requests));
  const identity = computed(() => {
    const { name, base_url, auth, inputs } = props.modelValue;
    return { name, base_url, auth, inputs };
  });
  const recipe = computed(() => ({ setup: props.modelValue.setup, refresh: props.modelValue.refresh }));
  function patch(value: Partial<ProviderDefinition>) {
    emit('update:modelValue', { ...props.modelValue, ...value });
  }
  function setRequest(id: string, value: ProviderRequest) {
    patch({ requests: { ...props.modelValue.requests, [id]: value } });
  }
  function removeRequest(id: string) {
    patch({ requests: Object.fromEntries(Object.entries(props.modelValue.requests).filter(([key]) => key !== id)) });
  }
  function addRequest() {
    setRequest(nextDefinitionKey(props.modelValue.requests, 'request'), {
      method: 'GET',
      path: ['v1'],
      query: {},
      captures: {},
    });
  }
</script>
<template>
  <div class="connection-form">
    <p class="connection-muted">
      These shared settings describe how the provider works. Your address and token are saved separately in Setup. Field
      locations use a slash, for example /collections; you only need to change them when the provider changes its
      response.
    </p>
    <ProviderIdentityFields :model-value="identity" @update:model-value="patch" />
    <h2>Requests</h2>
    <RouterLink :to="{ name: RouteName.IntegrationHelp, params: route.params, query: route.query, hash: '#values' }"
      >Help choosing request values</RouterLink
    >
    <p class="connection-muted">
      Each request has a name used by the setup and refresh steps below. Renaming it here updates those links.
    </p>
    <details v-for="(request, id) in modelValue.requests" :id="`requests.${id}`" :key="id" tabindex="-1">
      <summary>{{ humanLabel(String(id)) }} · {{ request.method }}</summary>
      <ProviderRequestFields
        :id="String(id)"
        :model-value="request"
        :choices="choices"
        @update:model-value="setRequest(String(id), $event)"
        @rename="emit('update:modelValue', renameProviderRequest(modelValue, String(id), $event))"
        @rename-capture="(oldName, name) => emit('update:modelValue', renameProviderCapture(modelValue, oldName, name))"
        @remove="removeRequest(String(id))" />
    </details>
    <div><UiButton @click="addRequest">Add request</UiButton></div>
    <p v-if="modelValue.integration_type === 'weather'" class="connection-muted">Edit weather selectors and location recipe in YAML mode.</p>
    <ProviderRecipeFields v-if="modelValue.integration_type !== 'weather'"
      :model-value="recipe"
      :request-names="requestNames"
      :inputs="modelValue.inputs"
      @update:model-value="patch" />
    <ProviderMappingFields v-if="'collections' in modelValue.mapping" :model-value="modelValue.mapping" @update:model-value="patch({ mapping: $event })" />
    <ProviderPolicyFields
      :model-value="{ policy: modelValue.policy, errors: modelValue.errors }"
      @update:model-value="patch" />
  </div>
</template>
