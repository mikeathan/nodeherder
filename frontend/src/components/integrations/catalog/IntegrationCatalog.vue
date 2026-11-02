<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import type { RouteLocationRaw } from 'vue-router';
  import type { ConnectionCatalog, ConnectionInstance } from '@/types/integrations';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import IntegrationTypeCard from './IntegrationTypeCard.vue';
  import {
    integrationTypeIcon,
    integrationTypeProviders,
    integrationTypeTitle,
    matchesCatalogQuery,
    providerCredentialLabel,
    providerRegionLabel,
    providerRefreshLabel,
  } from '@/domain/integrations/catalog';

  const props = defineProps<{
    catalog: ConnectionCatalog | null;
    hidden: ConnectionInstance[];
    busy: boolean;
    providerRoute: (key: string) => RouteLocationRaw;
  }>();
  defineEmits<{ add: [key: string]; restore: [id: string] }>();
  const selectedType = ref('');
  const query = ref('');
  const types = computed(() => {
    const catalog = props.catalog;
    return (catalog?.types ?? []).filter((type) =>
      matchesCatalogQuery(
        query.value,
        type.name,
        type.description,
        ...integrationTypeProviders(catalog, type.id).map((provider) => provider.name)
      )
    );
  });
  const providers = computed(() =>
    selectedType.value ? integrationTypeProviders(props.catalog, selectedType.value) : []
  );
  const activeTypeName = computed(() =>
    selectedType.value ? integrationTypeTitle(selectedType.value, props.catalog) : ''
  );
  watch(types, (current) => {
    if (selectedType.value && !current.some((type) => type.id === selectedType.value)) selectedType.value = '';
  });
  function providerCount(typeId: string): number {
    return integrationTypeProviders(props.catalog, typeId).length;
  }
</script>

<template>
  <div class="integration-catalog">
    <label class="catalog-search">
      <span>Search integrations</span>
      <input v-model="query" type="search" placeholder="Type or provider" autocomplete="off" />
    </label>

    <section v-if="hidden.length" aria-label="Hidden integrations">
      <h2>Hidden integrations</h2>
      <div v-for="item in hidden" :key="item.id" class="catalog-hidden-row">
        <div>
          <h3>{{ item.name }}</h3>
          <p>Configured and hidden from this page</p>
        </div>
        <UiButton :disabled="busy" @click="$emit('restore', item.id)">Show on page</UiButton>
      </div>
    </section>

    <section class="catalog-types" aria-label="Integration types">
      <IntegrationTypeCard
        v-for="type in types"
        :key="type.id"
        :type="type"
        :provider-count="providerCount(type.id)"
        :selected="selectedType === type.id"
        @select="selectedType = type.id" />
    </section>

    <p v-if="catalog?.available && !types.length" class="catalog-muted">
      {{ query ? 'No integrations match this search.' : 'No integration types are installed.' }}
    </p>
    <p v-if="catalog?.available && !catalog.providers.length" class="catalog-muted">
      No provider definitions are installed. Add provider files, then reload.
    </p>

    <section v-if="selectedType" class="catalog-providers" :aria-label="activeTypeName">
      <header>
        <h2>{{ activeTypeName }}</h2>
        <UiButton variant="ghost" :disabled="busy" @click="selectedType = ''">Close</UiButton>
      </header>
      <p v-if="!providers.length" class="catalog-muted">
        No provider is configured for this integration type. Add a provider definition before setting it up.
      </p>
      <p v-else-if="providers.length > 1" class="catalog-muted">
        Choose the provider for your region or data source. Coverage is shown where it is limited.
      </p>
      <div v-for="provider in providers" :key="provider.key" class="catalog-provider-row">
        <span class="catalog-provider-icon"><UiIcon :name="integrationTypeIcon(selectedType)" /></span>
        <div class="catalog-provider-copy">
          <div class="catalog-provider-heading">
            <h3>{{ provider.name }}</h3>
            <span v-if="providerRegionLabel(provider)" class="catalog-provider-region">
              {{ providerRegionLabel(provider) }}
            </span>
          </div>
          <p v-if="provider.disk_error" role="status">{{ provider.disk_error }}</p>
          <p>
            {{ providerCredentialLabel(provider) }} · {{ providerRefreshLabel(provider)
            }}<span v-if="provider.origin"> · {{ provider.origin }}</span>
          </p>
        </div>
        <div class="catalog-provider-actions">
          <UiButton
            variant="primary"
            :disabled="busy || !provider.available"
            @click="$emit('add', provider.key)"
            >Set up</UiButton
          >
          <UiButton :to="providerRoute(provider.key)" :disabled="busy">Configure operations</UiButton>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
  .integration-catalog {
    display: grid;
    gap: 1.5rem;
    min-width: 0;
  }
  .catalog-search {
    display: grid;
    gap: 0.4rem;
    font-weight: 600;
  }
  .catalog-search input {
    box-sizing: border-box;
    width: 100%;
    min-height: 44px;
    padding: 0.6rem 0.75rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
    color: var(--nh-text);
    font: inherit;
  }
  .catalog-types {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 18rem), 1fr));
    gap: 0.75rem;
  }
  .integration-catalog h2 {
    margin: 0 0 0.75rem;
    font-size: 1rem;
  }
  .integration-catalog h3 {
    margin: 0;
    font-size: 0.95rem;
    overflow-wrap: anywhere;
  }
  .catalog-provider-heading {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.45rem;
  }
  .catalog-provider-region {
    padding: 0.12rem 0.45rem;
    border-radius: 999px;
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
    font-size: 0.68rem;
    white-space: nowrap;
  }
  .integration-catalog p {
    margin: 0.3rem 0 0;
    color: var(--nh-text-muted);
    font-size: 0.82rem;
    overflow-wrap: anywhere;
  }
  .catalog-hidden-row,
  .catalog-provider-row {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.75rem;
    padding: 0.85rem 0;
    border-top: var(--nh-border-w) solid var(--nh-border);
  }
  .catalog-hidden-row > div,
  .catalog-provider-copy {
    flex: 1;
    min-width: 7rem;
  }
  .catalog-provider-icon {
    display: grid;
    place-items: center;
    flex: none;
    width: 2.2rem;
    height: 2.2rem;
    border-radius: 50%;
    background: var(--nh-surface-2);
    color: var(--nh-accent);
  }
  .catalog-provider-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }
  .catalog-providers header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }
  .catalog-muted {
    color: var(--nh-text-muted);
  }
</style>
