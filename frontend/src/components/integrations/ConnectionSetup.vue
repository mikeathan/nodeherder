<script setup lang="ts">
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import ConnectionFeedback from './ConnectionFeedback.vue';
  import { useConnectionSetup } from '@/composables/integrations/useConnectionSetup';
  import { RouteName } from '@/types/router';
  import './connections.css';
  const {
    reloadPage,
    setup,
    postcode,
    property,
    name,
    token,
    definition,
    busy,
    error,
    search,
    save,
    check,
    saveToken,
    cancel,
  } = useConnectionSetup();
</script>
<template>
  <div class="connections connection-stack">
    <UiPageHeader title="Set up bin collections" :back="{ name: RouteName.Connections }" back-label="Connections" />
    <ConnectionFeedback :error="error ?? setup?.last_error" :busy="busy" />
    <UiCard v-if="setup" title="Your address"
      ><div class="connection-form">
        <p class="connection-muted">
          The provider receives your postcode and selected address to retrieve collection dates. Refreshes are manual in
          this release.
        </p>
        <details v-if="definition?.auth.mode !== 'none' && definition?.auth.secret_slot">
          <summary>
            Provider credential
            {{
              setup.credential_configured ? '(saved)' : definition?.auth.mode === 'bearer' ? '(required)' : '(optional)'
            }}
          </summary>
          <p class="connection-muted">
            Use a token issued by the provider if your access requires one. It is stored privately and never included in
            the provider YAML.
          </p>
          <label class="connection-field"
            ><span>API token</span><input v-model="token" type="password" autocomplete="off" maxlength="8192" /></label
          ><UiButton :disabled="!token || busy" @click="saveToken">Save token</UiButton>
        </details>
        <form class="connection-form" @submit.prevent="search">
          <label class="connection-field"
            ><span>Postcode</span><input v-model="postcode" autocomplete="postal-code" maxlength="10" required
          /></label>
          <div><UiButton type="submit" :disabled="busy || !postcode.trim()">Find address</UiButton></div>
        </form>
        <form v-if="setup.status === 'choose'" class="connection-form" @submit.prevent="save">
          <label class="connection-field"
            ><span>Choose your address</span
            ><select v-model="property" required>
              <option disabled value="">Select an address</option>
              <option v-for="choice in setup.choices" :key="choice.id" :value="choice.id">{{ choice.label }}</option>
            </select></label
          >
          <p v-if="!setup.choices.length" class="connection-muted">
            No addresses were found. Check the postcode or try another provider.
          </p>
          <label class="connection-field"
            ><span>Connection name</span><input v-model="name" maxlength="120" required
          /></label>
          <div>
            <UiButton type="submit" variant="primary" :disabled="busy || !property || !name.trim()"
              >Save connection</UiButton
            >
          </div>
        </form>
        <div v-if="['searching', 'saving', 'pending', 'error'].includes(setup.status) && setup.operation_id">
          <p class="connection-muted">
            {{
              setup.status === 'searching'
                ? 'Finding addresses…'
                : 'The provider has not finished. Use Check status to continue the same lookup.'
            }}
          </p>
          <UiButton :disabled="busy" @click="check">Check status</UiButton>
        </div>
        <UiButton v-if="setup.instance_id" :to="{ name: RouteName.Connection, params: { id: setup.instance_id } }"
          >Open saved connection</UiButton
        >
        <div><UiButton variant="ghost" :disabled="busy" @click="cancel">Discard setup</UiButton></div>
      </div></UiCard
    >
    <UiButton v-if="error" @click="reloadPage">Reload setup</UiButton>
  </div>
</template>
