<script setup lang="ts">
  /*
   * New automation: pick the device whose changes start it. Automations are keyed by that
   * device, so a device that already has one opens its existing automation instead.
   */
  import { computed, ref } from 'vue';
  import { RouteName } from '@/types/router';
  import { useHub } from '@/composables/useHub';
  import { useAutomations } from '@/composables/useAutomations';
  import { deviceIcon } from '@/components/ui/icons';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import UiChip from '@/components/ui/UiChip.vue';
  import UiEmpty from '@/components/ui/UiEmpty.vue';

  const { devices } = useHub();
  const { find } = useAutomations();
  const query = ref('');
  const rows = computed(() => {
    const q = query.value.trim().toLowerCase();
    return devices.value.filter((d) => !q || `${d.friendly_name} ${d.description} ${d.id}`.toLowerCase().includes(q)).map((d) => ({ device: d, existing: !!find(d.id) }));
  });
</script>

<template>
  <div class="nh-creator">
    <UiPageHeader title="New automation" subtitle="Which device should start it?" :back="{ name: RouteName.Viewer }" back-label="Automations" />
    <label class="nh-creator-search">
      <UiIcon name="search" />
      <span class="sr-only">Search devices</span>
      <input v-model="query" class="nh-input" type="search" placeholder="Search devices" />
    </label>
    <UiEmpty v-if="!rows.length" icon="search" title="No devices match" />
    <ul v-else class="nh-pick">
      <li v-for="r in rows" :key="r.device.id">
        <RouterLink :to="{ name: RouteName.Editor, params: { id: r.device.id } }" class="nh-pick-row">
          <span class="nh-pick-ic"><UiIcon :path="deviceIcon(r.device)" /></span>
          <span class="nh-pick-t"><b>{{ r.device.friendly_name }}</b><small>{{ r.device.description || r.device.id }}</small></span>
          <UiChip v-if="r.existing" tone="accent" icon="edit">Has one · edit</UiChip>
          <UiIcon name="chevronRight" class="nh-muted" />
        </RouterLink>
      </li>
    </ul>
  </div>
</template>

<style scoped>
  .nh-creator { max-width: 44rem; }
  .nh-creator-search { position: relative; display: flex; align-items: center; margin-bottom: var(--nh-space); }
  .nh-creator-search > .nh-ic { position: absolute; left: 0.65rem; color: var(--nh-text-muted); }
  .nh-creator-search .nh-input { width: 100%; padding-left: 2.2rem; }
  .nh-pick { list-style: none; margin: 0; padding: 0; background: var(--nh-surface); border-radius: var(--nh-radius-l); box-shadow: var(--nh-shadow-1); overflow: hidden; }
  .nh-pick li + li { border-top: var(--nh-border-w) solid var(--nh-border); }
  .nh-pick-row { display: flex; align-items: center; gap: 0.75rem; padding: 0.65rem var(--nh-space); text-decoration: none; min-height: 3.5rem; }
  .nh-pick-row:hover { background: var(--nh-surface-2); }
  .nh-pick-ic { display: grid; place-items: center; width: 2.2rem; height: 2.2rem; border-radius: var(--nh-radius-m); background: var(--nh-accent-soft); color: var(--nh-accent); flex: none; }
  .nh-pick-t { flex: 1; display: flex; flex-direction: column; min-width: 0; }
  .nh-pick-t b, .nh-pick-t small { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
  .nh-pick-t small { color: var(--nh-text-muted); }
</style>
