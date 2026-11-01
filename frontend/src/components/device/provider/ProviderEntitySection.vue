<script setup lang="ts">
  /* One searchable, alphabetically sorted page of provider entities. */
  import type { Device, Expose } from '@/types/device';
  import { showsExposeDescription } from '@/domain/exposes';
  import type { IconName } from '@/components/ui/icons';
  import { useProviderEntityInventory } from '@/composables/device/useProviderEntityInventory';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import ExposeRow from '@/components/entity/ExposeRow.vue';

  const props = defineProps<{ device: Device; entities: readonly Expose[]; title: string; icon: IconName; showDescription?: boolean }>();
  const { query, page, pages, filtered, visible, pageSize } = useProviderEntityInventory(
    () => props.entities,
    () => props.device.id
  );
</script>

<template>
  <UiCard :title="title" :icon="icon" flush>
    <div class="nh-provider-search">
      <label :for="`${device.id}-${icon}-search`">Search {{ title.toLowerCase() }}</label>
      <input :id="`${device.id}-${icon}-search`" v-model="query" class="nh-input" type="search" />
    </div>
    <ExposeRow
      v-for="entity in visible"
      :key="entity.name"
      :device="device"
      :expose="entity"
      :show-description="showsExposeDescription(entity, showDescription)" />
    <p v-if="!filtered.length" class="nh-card-body nh-muted">{{ entities.length ? 'No matching entities.' : 'This device reports no readings.' }}</p>
    <div v-if="filtered.length > pageSize" class="nh-provider-pages" :aria-label="`${title} pages`">
      <UiButton variant="ghost" :disabled="page <= 1" :aria-label="`Previous ${title.toLowerCase()} page`" @click="page--">Previous</UiButton>
      <span aria-live="polite">Page {{ page }} of {{ pages }} · {{ filtered.length }} items</span>
      <UiButton variant="ghost" :disabled="page >= pages" :aria-label="`Next ${title.toLowerCase()} page`" @click="page++">Next</UiButton>
    </div>
  </UiCard>
</template>

<style scoped>
  .nh-provider-search {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.5rem;
    padding: var(--nh-space);
  }
  .nh-provider-search input {
    flex: 1 1 10rem;
    min-width: 0;
  }
  .nh-provider-pages {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
    padding: var(--nh-space);
  }
</style>
