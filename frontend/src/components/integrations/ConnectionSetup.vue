<script setup lang="ts">
  import { computed } from 'vue';
  import { setupPresentation } from '@/integrations/registry';
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
  const presentation = computed(() => (definition.value ? setupPresentation(definition.value) : null));
</script>
<template>
  <div class="connections connection-stack">
    <UiPageHeader
      :title="presentation?.title ?? 'Set up integration'"
      :back="{ name: RouteName.Integrations }"
      back-label="Integrations" />
    <ConnectionFeedback :error="error ?? setup?.last_error" :busy="busy" />
    <UiCard v-if="setup" :title="presentation?.sectionTitle ?? 'Your selection'"
      ><div class="connection-form">
        <p class="connection-muted">{{ presentation?.notice }}</p>
        <details v-if="definition?.auth.mode !== 'none' && definition?.auth.secret_slot">
          <summary>
            Provider credential
            {{
              setup.credential_configured
                ? '(saved)'
                : definition?.auth.mode === 'optional_bearer'
                  ? '(optional)'
                  : '(required)'
            }}
          </summary>
          <p class="connection-muted">
            Use a token issued by the provider if your access requires one. It is stored privately and never included in
            the provider YAML.
          </p>
          <label class="connection-field"
            ><span>API key / token</span><input v-model="token" type="password" autocomplete="off" maxlength="8192"
          /></label>
          <p v-if="presentation?.keyURL">
            <a :href="presentation?.keyURL" target="_blank" rel="noopener noreferrer"
              >Get an API key from the provider ↗</a
            >
          </p>
          <UiButton :disabled="!token || busy" @click="saveToken">Save token</UiButton>
        </details>
        <form class="connection-form" @submit.prevent="search">
          <label class="connection-field"
            ><span>{{ presentation?.searchLabel }}</span
            ><input
              v-model="postcode"
              :autocomplete="presentation?.autocomplete"
              :maxlength="presentation?.searchMaxLength"
              required
          /></label>
          <div>
            <UiButton type="submit" :disabled="busy || !postcode.trim()">{{ presentation?.searchButton }}</UiButton>
          </div>
        </form>
        <form v-if="setup.status === 'choose'" class="connection-form" @submit.prevent="save">
          <label class="connection-field"
            ><span>{{ presentation?.choiceLabel }}</span
            ><select v-model="property" required>
              <option disabled value="">{{ presentation?.choicePlaceholder }}</option>
              <option v-for="choice in setup.choices" :key="choice.id" :value="choice.id">{{ choice.label }}</option>
            </select></label
          >
          <p v-if="!setup.choices.length" class="connection-muted">
            No results were found. Check the postcode or try another provider.
          </p>
          <label class="connection-field"
            ><span>Integration name</span><input v-model="name" maxlength="120" required
          /></label>
          <div>
            <UiButton type="submit" variant="primary" :disabled="busy || !property || !name.trim()"
              >Save integration</UiButton
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
        <UiButton v-if="setup.instance_id" :to="{ name: RouteName.Integration, params: { id: setup.instance_id } }"
          >Open saved integration</UiButton
        >
        <div><UiButton variant="ghost" :disabled="busy" @click="cancel">Discard setup</UiButton></div>
      </div></UiCard
    >
    <UiButton v-if="error" @click="reloadPage">Reload setup</UiButton>
  </div>
</template>
