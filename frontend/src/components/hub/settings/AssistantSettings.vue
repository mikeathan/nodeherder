<script setup lang="ts">
  /*
   * Assistant settings (spec 059 FR-03): the service address the hub forwards chat
   * messages to. Saving builds a new settings object and validates the URL before any
   * frame is sent, so an invalid address never reaches the hub.
   */
  import { computed, ref, watch } from 'vue';
  import { store } from '@/store';
  import type { AssistantSettingsType } from '@/types/settings.type';
  import { ASSISTANT_URL_ERROR, withAssistantUrl } from '@/domain/settings/policy';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiField from '@/components/ui/UiField.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const settings = computed<AssistantSettingsType | undefined>(() => store.getters['hub/assistant']());
  const offline = computed(() => store.getters['ws/getConnectionStatus'] !== 'connected');

  const draft = ref(settings.value?.url ?? '');
  const error = ref('');
  const saved = ref(false);

  // A hub update (another tab, reconnect) replaces the draft unless the user is mid-edit.
  watch(
    () => settings.value?.url,
    (url) => {
      if (error.value) return;
      draft.value = url ?? '';
    }
  );

  function save() {
    const next = withAssistantUrl(settings.value, draft.value);
    if (!next) {
      error.value = ASSISTANT_URL_ERROR;
      return;
    }
    error.value = '';
    if (next.url === (settings.value?.url ?? '')) return;
    store.dispatch('hub/saveAssistantSettings', next);
    saved.value = true;
  }
</script>

<template>
  <section class="nh-section" aria-labelledby="assistant-heading">
    <header class="nh-section-head">
      <h2 id="assistant-heading">Assistant</h2>
      <p class="nh-help">Chat messages are sent to this address. The Assistant screen uses it for every request.</p>
    </header>

    <p v-if="offline" class="nh-alert is-warn" role="status">
      <UiIcon name="offline" />
      <span>Connect to the hub to change the assistant address.</span>
    </p>

    <UiCard title="Service address" icon="settingsAssistant" :heading-level="3">
      <form class="nh-stack" @submit.prevent="save">
        <UiField
          label="Assistant URL"
          for="assistant-url"
          help="Absolute http:// or https:// address. Leave empty to clear it."
          :error="error || null">
          <input
            id="assistant-url"
            v-model="draft"
            class="nh-input"
            type="url"
            inputmode="url"
            autocomplete="off"
            spellcheck="false"
            placeholder="http://hub.local:4001/api/conversation"
            :disabled="offline"
            :aria-invalid="!!error || undefined"
            :aria-describedby="error ? 'assistant-url-err' : 'assistant-url-help'"
            @input="saved = false" />
        </UiField>
        <div class="nh-inline">
          <UiButton type="submit" variant="primary" icon="check" :disabled="offline">Save address</UiButton>
          <span v-if="saved && !error" class="nh-muted" role="status">
            <UiIcon name="check" /> Address sent to the hub.
          </span>
        </div>
      </form>
    </UiCard>
  </section>
</template>
