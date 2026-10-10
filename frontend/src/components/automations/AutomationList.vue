<script setup lang="ts">
  /* Automations list (spec 007 US-04): summary sentence, on/off, run, edit and delete. */
  import { computed, ref } from 'vue';
  import { RouteName } from '@/types/router';
  import { Automation } from '@/types/automation.type';
  import { canTriggerManually } from '@/contracts/automations';
  import { describeTrigger } from '@/domain/automation';
  import { useAutomations } from '@/composables/useAutomations';
  import { findDevice } from '@/composables/useHub';
  import { confirm } from '@/composables/useConfirm';
  import { deviceIcon, ICONS } from '@/components/ui/icons';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiEmpty from '@/components/ui/UiEmpty.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import UiToggle from '@/components/ui/UiToggle.vue';
  import UiChip from '@/components/ui/UiChip.vue';

  const { automations, loaded, setEnabled, remove, run } = useAutomations();
  const query = ref('');

  const rows = computed(() => {
    const q = query.value.trim().toLowerCase();
    return [...automations.value]
      .filter((a) => !q || `${a.friendlyname} ${a.description}`.toLowerCase().includes(q))
      .sort((a, b) => a.friendlyname.localeCompare(b.friendlyname, undefined, { sensitivity: 'base' }))
      .map((a) => {
        const device = findDevice(a.id);
        const runnable = a.triggers?.length === 1 && canTriggerManually(a.triggers[0]) ? a.triggers[0] : null;
        return {
          automation: a,
          icon: device ? deviceIcon(device) : ICONS.automation,
          deviceName: device?.friendly_name ?? 'Device not found',
          summary: describeTrigger(a, a.triggers?.[0], findDevice),
          more: Math.max(0, (a.triggers?.length ?? 0) - 1),
          scheduled: (a.schedules?.length ?? 0) > 0,
          runnable,
        };
      });
  });

  async function onDelete(a: Automation) {
    const ok = await confirm({ title: `Delete “${a.friendlyname}”?`, message: 'The automation and all its triggers are removed from the hub.', confirmLabel: 'Delete', danger: true });
    if (ok) remove(a.id);
  }
</script>

<template>
  <div>
    <UiPageHeader title="Automations" :subtitle="`${automations.length} automations · ${automations.filter((a) => a.enabled).length} on`">
      <template #actions>
        <UiButton variant="primary" icon="add" :to="{ name: RouteName.Creator }">New automation</UiButton>
      </template>
    </UiPageHeader>

    <label v-if="automations.length > 4" class="nh-auto-search">
      <UiIcon name="search" />
      <span class="sr-only">Search automations</span>
      <input v-model="query" class="nh-input" type="search" placeholder="Search automations" />
    </label>

    <div v-if="!loaded && !automations.length" class="nh-auto-list" aria-busy="true">
      <div v-for="n in 3" :key="n" class="nh-card nh-skel-card"><span class="nh-skel" /><span class="nh-skel" /></div>
    </div>
    <UiEmpty v-else-if="!automations.length" icon="automation" title="No automations yet" text="Automations react to a device change, check conditions, then control other devices.">
      <UiButton variant="primary" icon="add" :to="{ name: RouteName.Creator }">Create an automation</UiButton>
    </UiEmpty>
    <UiEmpty v-else-if="!rows.length" icon="search" title="No automations match" />
    <ul v-else class="nh-auto-list">
      <li v-for="r in rows" :key="r.automation.id" class="nh-auto" :data-automation="r.automation.id">
        <RouterLink :to="{ name: RouteName.Editor, params: { id: r.automation.id } }" class="nh-auto-main">
          <span class="nh-auto-ic"><UiIcon :path="r.icon" /></span>
          <span class="nh-auto-t">
            <b class="nh-auto-name">{{ r.automation.friendlyname }}</b>
            <small>{{ r.automation.description || r.deviceName }}</small>
            <span class="nh-auto-sum">{{ r.summary }}<em v-if="r.more"> +{{ r.more }} more trigger{{ r.more === 1 ? '' : 's' }}</em></span>
          </span>
        </RouterLink>
        <div class="nh-auto-side">
          <UiChip v-if="r.scheduled" icon="schedule" tone="info">Scheduled</UiChip>
          <UiButton v-if="r.runnable" size="sm" icon="run" @click="run(r.automation, r.runnable.name)">Run</UiButton>
          <UiToggle :model-value="r.automation.enabled" :label="`${r.automation.friendlyname} on`" @update:model-value="setEnabled(r.automation, $event)" />
          <UiButton size="sm" variant="ghost" icon="delete" class="is-danger" :label="`Delete ${r.automation.friendlyname}`" @click="onDelete(r.automation)" />
        </div>
      </li>
    </ul>
  </div>
</template>

<style scoped>
  .nh-auto-search { position: relative; display: flex; align-items: center; max-width: 26rem; margin-bottom: var(--nh-space); }
  .nh-auto-search > .nh-ic { position: absolute; left: 0.65rem; color: var(--nh-text-muted); }
  .nh-auto-search .nh-input { width: 100%; padding-left: 2.2rem; }
  .nh-auto-list { list-style: none; margin: 0; padding: 0; display: flex; flex-direction: column; gap: var(--nh-gap); }
  .nh-auto { display: flex; flex-wrap: wrap; align-items: center; gap: 0.75rem; padding: 0.8rem var(--nh-space); background: var(--nh-surface); border: var(--nh-border-w) solid transparent; border-radius: var(--nh-radius-l); box-shadow: var(--nh-shadow-1); }
  [data-mode='dark'] .nh-auto { border-color: var(--nh-border); }
  .nh-auto:hover { border-color: var(--nh-border-strong); }
  .nh-auto-main { flex: 1 1 22rem; display: flex; gap: 0.8rem; align-items: flex-start; text-decoration: none; min-width: 0; }
  .nh-auto-main:hover .nh-auto-name { color: var(--nh-accent); }
  .nh-auto-ic { display: grid; place-items: center; width: 2.3rem; height: 2.3rem; border-radius: var(--nh-radius-m); background: var(--nh-accent-soft); color: var(--nh-accent); flex: none; }
  .nh-auto-t { display: flex; flex-direction: column; min-width: 0; }
  .nh-auto-t small { color: var(--nh-text-muted); }
  .nh-auto-sum { margin-top: 0.3rem; font-size: 0.82rem; color: var(--nh-text-muted); }
  .nh-auto-sum em { color: var(--nh-accent); font-style: normal; }
  .nh-auto-side { display: flex; flex-wrap: wrap; align-items: center; gap: 0.5rem; margin-left: auto; }
  .nh-skel-card { display: flex; flex-direction: column; gap: 0.6rem; padding: 1rem; }
  @media (max-width: 640px) {
    .nh-auto-side { flex-basis: 100%; justify-content: flex-end; padding-top: 0.5rem; border-top: var(--nh-border-w) solid var(--nh-border); }
  }
</style>
