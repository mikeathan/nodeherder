<script setup lang="ts">
  import { onMounted } from 'vue';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import ConnectionFeedback from './ConnectionFeedback.vue';
  import { useConnections } from '@/composables/useConnections';
  import { connectionsApi } from '@/services/integrations.service';
  import { RouteName } from '@/types/router';
  import './connections.css';
  const { state, store, router, busy, error, perform, signal } = useConnections();
  const load = () => perform(() => store.dispatch('integrations/load', signal));
  onMounted(load);
  async function add(key: string) {
    const setup = await perform(() => connectionsApi.createSetup(key));
    if (setup) await router.push({ name: RouteName.ConnectionSetup, params: { id: setup.id } });
  }
  async function reloadProviders() {
    const value = await perform(() => connectionsApi.reloadProviders());
    if (value) await load();
  }
  const providerRoute = (key: string) => {
    const [type, provider] = key.split('/');
    return { name: RouteName.ConnectionProvider, params: { type, provider } };
  };
</script>
<template>
  <div class="connections connection-stack">
    <UiPageHeader title="Connections" subtitle="Services and information connected to NodeHerder."
      ><template #actions
        ><UiButton :disabled="busy || !state.catalog?.available" @click="reloadProviders"
          >Reload providers</UiButton
        ></template
      ></UiPageHeader
    >
    <ConnectionFeedback :error="error" :busy="busy" />
    <div v-if="state.catalog?.availability_error" class="connection-notice" role="status">
      <p>{{ state.catalog.availability_error.message }}</p>
      <p class="connection-muted">
        The private key must be mounted separately from the saved configuration. Existing device features remain
        available.
      </p>
      <UiButton @click="load">Check again</UiButton>
    </div>
    <section aria-labelledby="saved-title">
      <h2 id="saved-title">Your connections</h2>
      <p v-if="!state.instances.length && state.catalog?.available" class="connection-muted">
        No connections yet. Choose a service below to get started.
      </p>
      <div class="connection-grid">
        <UiCard v-for="item in state.instances" :key="item.id" :title="item.name" icon="schedule"
          ><p>{{ item.postcode }}</p>
          <p class="connection-muted">
            {{
              !item.enabled
                ? 'Disabled'
                : item.last_error
                  ? item.last_error.message
                  : item.resources.length
                    ? 'Collection dates saved'
                    : 'Waiting for collection dates'
            }}
          </p>
          <UiButton :to="{ name: RouteName.Connection, params: { id: item.id } }">Open connection</UiButton></UiCard
        >
      </div>
    </section>
    <section v-for="type in state.catalog?.types" :key="type.id" :aria-labelledby="`type-${type.id}`">
      <h2 :id="`type-${type.id}`">{{ type.name }}</h2>
      <p class="connection-muted">{{ type.description }}</p>
      <div class="connection-grid">
        <UiCard
          v-for="provider in state.catalog?.providers.filter((p) => p.integration_type === type.id)"
          :key="provider.key"
          :title="provider.name">
          <p class="connection-muted connection-break">{{ provider.origin }}</p>
          <p v-if="provider.disk_error">{{ provider.disk_error }}</p>
          <div class="connection-actions">
            <UiButton variant="primary" :disabled="!provider.available || busy" @click="add(provider.key)"
              >Add connection</UiButton
            ><UiButton :to="providerRoute(provider.key)">Configure operations</UiButton>
          </div>
        </UiCard>
      </div>
    </section>
  </div>
</template>
