<script setup lang="ts">
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import type { ConnectionInstance, ProviderDefinition } from '@/types/integrations.type';
  defineProps<{ item: ConnectionInstance; definition: ProviderDefinition | null; busy: boolean }>();
  const name = defineModel<string>('name', { required: true });
  const token = defineModel<string>('token', { required: true });
  const emit = defineEmits<{
    update: [enabled: boolean];
    editAddress: [];
    credential: [remove: boolean, slot?: string];
    remove: [];
  }>();
</script>
<template>
  <UiCard title="Integration settings"
    ><div class="connection-form">
      <label class="connection-field"><span>Integration name</span><input v-model="name" maxlength="120" /></label>
      <div class="connection-actions">
        <UiButton :disabled="busy || !name.trim()" @click="emit('update', item.enabled)">Save name</UiButton
        ><UiButton :disabled="busy" @click="emit('update', !item.enabled)">{{
          item.enabled ? 'Disable integration' : 'Enable integration'
        }}</UiButton
        ><UiButton :disabled="busy" @click="emit('editAddress')">Change address</UiButton>
      </div>
      <p class="connection-muted">
        Changing the address uses a separate setup. Your saved dates remain available until the new address is verified.
      </p>
      <fieldset v-if="definition?.auth.secret_slot || item.credential_slots.length">
        <legend>Provider token</legend>
        <p class="connection-muted">
          {{ item.credential_configured ? 'A token is saved. Its value cannot be displayed.' : 'No token is saved.' }}
        </p>
        <p class="connection-muted">
          Saved tokens for other slot names are not used by the current configuration. Removing a token stops any
          pending attempt.
        </p>
        <label v-if="definition?.auth.mode !== 'none' && definition?.auth.secret_slot" class="connection-field"
          ><span>{{
            item.credential_slots.includes(definition.auth.secret_slot) ? 'Replacement API token' : 'API token'
          }}</span
          ><input v-model="token" type="password" autocomplete="off" maxlength="8192"
        /></label>
        <div class="connection-actions">
          <UiButton
            v-if="definition?.auth.mode !== 'none' && definition?.auth.secret_slot"
            :disabled="busy || !token"
            @click="emit('credential', false)"
            >Save token</UiButton
          ><UiButton
            v-for="slot in item.credential_slots"
            :key="slot"
            :disabled="busy"
            @click="emit('credential', true, slot)"
            >Remove saved token ({{ slot }})</UiButton
          >
        </div>
      </fieldset>
      <div class="connection-delete">
        <UiButton variant="danger" :disabled="busy" @click="emit('remove')">Remove integration</UiButton>
      </div>
    </div></UiCard
  >
</template>
