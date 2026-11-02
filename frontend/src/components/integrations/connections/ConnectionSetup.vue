<script setup lang="ts">
  import { computed } from 'vue';
  import { setupPresentation } from '@/integrations/registry';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import ConnectionFeedback from '../shared/ConnectionFeedback.vue';
  import { useConnectionSetup } from '@/composables/integrations/useConnectionSetup';
  import { credentialLabel, credentialSlots } from '@/domain/integrations/credentials';
  import { formatScheduleInterval } from '@/domain/schedule';
  import { RouteName } from '@/types/router';
  import '../connections.css';
  const {
    reloadPage,
    setup,
    postcode,
    property,
    name,
    tokens,
    definition,
    busy,
    error,
    search,
    save,
    check,
    saveCredentials,
    cancel,
  } = useConnectionSetup();
  const presentation = computed(() => (definition.value ? setupPresentation(definition.value) : null));
  const slots = computed(() => credentialSlots(definition.value?.auth));
  const providerHost = computed(() => {
    const raw = definition.value?.base_url ?? '';
    try {
      return new URL(raw).host;
    } catch {
      return raw;
    }
  });
  const providerLink = computed(() => presentation.value?.keyURL ?? presentation.value?.providerURL ?? definition.value?.base_url);
  const providerLinkLabel = computed(() => {
    if (!presentation.value?.keyURL) return 'Open provider site';
    return definition.value?.auth.mode === 'none' ? 'Provider documentation' : 'Get credentials';
  });
  const credentialSummary = computed(() => {
    switch (definition.value?.auth.mode) {
      case 'none':
        return 'No credential required';
      case 'optional_bearer':
        return 'Optional credential';
      case 'oauth2_client_credentials':
        return 'OAuth client credentials required';
      default:
        return 'Credential required';
    }
  });
  const refreshSummary = computed(() => {
    const policy = definition.value?.policy;
    if (!policy?.refresh_default_seconds) return 'Manual refresh';
    return `Refreshes ${formatScheduleInterval(policy.refresh_default_seconds)} · minimum ${formatScheduleInterval(policy.refresh_min_seconds)}`;
  });
  const hasCredentialInput = computed(() => slots.value.some((slot) => Boolean(tokens.value[slot])));
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
        <section class="setup-provider" aria-label="Provider">
          <div class="setup-provider-main">
            <span class="setup-provider-label">Provider</span>
            <strong>{{ definition?.name ?? 'Provider' }}</strong>
            <span class="setup-provider-host">{{ providerHost }}</span>
          </div>
          <div class="setup-provider-meta">
            <span>{{ credentialSummary }}</span>
            <span>{{ refreshSummary }}</span>
          </div>
          <div class="setup-provider-links">
            <a v-if="providerLink" :href="providerLink" target="_blank" rel="noopener noreferrer"
              >{{ providerLinkLabel }} ↗</a
            >
            <a
              v-for="service in presentation?.services"
              :key="service.url"
              :href="service.url"
              target="_blank"
              rel="noopener noreferrer"
              >{{ service.label }} ↗</a
            >
          </div>
        </section>
        <p class="connection-muted">{{ presentation?.notice }}</p>
        <section v-if="slots.length" class="setup-credentials" aria-labelledby="setup-credentials-heading">
          <div class="setup-credentials-head">
            <h3 id="setup-credentials-heading">Provider credentials</h3>
            <span>
              {{
                setup.credential_configured
                  ? 'Saved'
                  : definition?.auth.mode === 'optional_bearer'
                    ? 'Optional'
                    : 'Required'
              }}
            </span>
          </div>
          <p class="connection-muted">
            Credentials are stored privately in their named slots and never included in the provider YAML.
          </p>
          <label v-for="slot in slots" :key="slot" class="connection-field" :for="`credential-${slot}`">
            <span>
              {{ credentialLabel(definition?.auth, slot) }}
              <em v-if="setup.credential_slots?.includes(slot)" class="setup-credential-saved">Saved</em>
            </span>
            <input
              :id="`credential-${slot}`"
              v-model="tokens[slot]"
              type="password"
              autocomplete="off"
              maxlength="8192" />
          </label>
          <div>
            <UiButton :disabled="busy || !hasCredentialInput" @click="saveCredentials(slots)"
              >Save credentials</UiButton
            >
          </div>
        </section>
        <form class="connection-form" @submit.prevent="search">
          <label class="connection-field"
            ><span>{{ presentation?.searchLabel }}</span
            ><input
              v-model="postcode"
              :autocomplete="presentation?.autocomplete"
              :maxlength="presentation?.searchMaxLength"
              :placeholder="presentation?.searchPlaceholder"
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
            {{ presentation?.noResults ?? 'No results were found. Check the search or try another provider.' }}
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
                : 'The provider is still collecting data. This page checks for up to five minutes; then use Check status.'
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

<style scoped>
  .setup-provider {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem 1.5rem;
    padding: 0.9rem 1rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface-2);
  }
  .setup-provider-main {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    min-width: 10rem;
  }
  .setup-provider-label {
    color: var(--nh-text-muted);
    font-size: 0.68rem;
    font-weight: 600;
    letter-spacing: 0.04em;
    text-transform: uppercase;
  }
  .setup-provider-main strong {
    font-size: 1rem;
  }
  .setup-provider-host,
  .setup-provider-meta {
    color: var(--nh-text-muted);
    font-size: 0.8rem;
    overflow-wrap: anywhere;
  }
  .setup-provider-meta {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    margin-inline-start: auto;
  }
  .setup-provider-links {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem 0.9rem;
    width: 100%;
  }
  .setup-provider-links a {
    font-size: 0.8rem;
    font-weight: 600;
  }
  .setup-credentials {
    display: grid;
    gap: 0.7rem;
    padding: 0.9rem 1rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
  }
  .setup-credentials-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 1rem;
  }
  .setup-credentials-head h3 {
    margin: 0;
    font-size: 0.9rem;
  }
  .setup-credentials-head > span {
    color: var(--nh-text-muted);
    font-size: 0.75rem;
  }
  .setup-credential-saved {
    margin-inline-start: 0.4rem;
    color: var(--nh-ok);
    font-size: 0.7rem;
    font-style: normal;
    font-weight: 600;
  }
  @media (max-width: 640px) {
    .setup-provider-meta {
      margin-inline-start: 0;
    }
  }
</style>
