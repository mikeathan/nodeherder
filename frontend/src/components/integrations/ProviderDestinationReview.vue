<script setup lang="ts">
  import type { CredentialReview, DefinitionValidation } from '@/types/integrations.type';
  defineProps<{
    validation: DefinitionValidation;
    instances: (CredentialReview & { name: string })[];
    busy: boolean;
  }>();
  const approved = defineModel<boolean>('approved', { required: true });
  const reassociate = defineModel<string[]>('reassociate', { required: true });
</script>
<template>
  <fieldset class="connection-form" :disabled="busy">
    <legend>Provider destination changed</legend>
    <p class="connection-break">From {{ validation.previous_origin }} to {{ validation.origin }}.</p>
    <label><input v-model="approved" type="checkbox" /> I approve this destination for future requests.</label>
    <p class="connection-muted">
      Choose which existing integrations may send their saved inputs and token to it. Others will need reconnecting.
    </p>
    <label v-for="item in instances" :key="item.instance_id">
      <input v-model="reassociate" type="checkbox" :value="item.instance_id" />
      {{ item.name }} {{ item.credential_revision ? '— includes its saved token' : '— saved inputs only' }}
    </label>
  </fieldset>
</template>
