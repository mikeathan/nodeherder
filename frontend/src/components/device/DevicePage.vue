<script setup lang="ts">
  /*
   * Device page (spec 007 US-03): controls and readings, history, local settings and details.
   * The tab is kept in the URL (?tab=) so it survives reloads and can be linked.
   * Back returns to where the user came from, or to the device list when opened directly.
   */
  import { computed, defineAsyncComponent } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { RouteName } from '@/types/router';
  import { ExposeCategories } from '@/types/device.type';
  import { useHub } from '@/composables/useHub';
  import { resolveProtocol } from '@/domain/devices';
  import { deviceIcon } from '@/components/ui/icons';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiEmpty from '@/components/ui/UiEmpty.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import RelativeTime from '@/components/ui/RelativeTime.vue';
  import ExposeRow from '@/components/entity/ExposeRow.vue';
  import DeviceChips from '@/components/entity/DeviceChips.vue';
  import DeviceStatusChip from '@/components/entity/DeviceStatusChip.vue';
  import DeviceActions from './DeviceActions.vue';

  const DeviceMetrics = defineAsyncComponent(() => import('./DeviceMetrics.vue'));
  const DeviceSettings = defineAsyncComponent(() => import('./DeviceSettings.vue'));

  const props = defineProps<{ id: string }>();
  const route = useRoute();
  const router = useRouter();
  const { findDevice, initialized, statusOf } = useHub();

  const device = computed(() => findDevice(props.id));
  const status = computed(() => (device.value ? statusOf(device.value) : 'unknown'));

  const TABS = [
    { id: 'controls', label: 'Controls' },
    { id: 'history', label: 'History' },
    { id: 'settings', label: 'Settings' },
    { id: 'about', label: 'About' },
  ] as const;
  type TabId = (typeof TABS)[number]['id'];
  const tab = computed<TabId>(() => (TABS.some((t) => t.id === route.query.tab) ? (route.query.tab as TabId) : 'controls'));
  const selectTab = (id: TabId) => router.replace({ query: { ...route.query, tab: id === 'controls' ? undefined : id } });
  function onTabKey(e: KeyboardEvent, index: number) {
    const step = e.key === 'ArrowRight' ? 1 : e.key === 'ArrowLeft' ? -1 : 0;
    if (!step) return;
    const next = TABS[(index + step + TABS.length) % TABS.length];
    selectTab(next.id);
    (document.getElementById(`tab-${next.id}`) as HTMLElement | null)?.focus();
  }

  const back = computed(() => {
    const previous = router.options.history.state.back;
    return typeof previous === 'string' ? previous : { name: RouteName.DeviceList };
  });

  const byCategory = (category: string) => Object.values(device.value?.exposes ?? {}).filter((e) => e.category === category);
  const measurements = computed(() => byCategory(ExposeCategories.Measurement));
  const configuration = computed(() => byCategory(ExposeCategories.Config));
  const diagnostics = computed(() => byCategory(ExposeCategories.Diagnostic));
  const protocol = computed(() => resolveProtocol(device.value?.connection_type));
</script>

<template>
  <div v-if="device">
    <UiPageHeader :title="device.friendly_name" :subtitle="device.description || undefined" :back="back" back-label="Back">
      <template #before>
        <span class="nh-dev-hero"><UiIcon :path="deviceIcon(device)" /></span>
      </template>
      <template #actions>
        <DeviceActions :device="device" />
      </template>
    </UiPageHeader>
    <div class="nh-dev-chips">
      <DeviceStatusChip :device="device" />
      <DeviceChips :device="device" />
      <span class="nh-muted nh-dev-seen">Last seen <RelativeTime :value="device.last_seen" /></span>
    </div>

    <div class="nh-tabs" role="tablist" aria-label="Device sections">
      <button
        v-for="(t, i) in TABS"
        :id="`tab-${t.id}`"
        :key="t.id"
        type="button"
        role="tab"
        class="nh-tab"
        :aria-selected="tab === t.id"
        :aria-controls="`panel-${t.id}`"
        :tabindex="tab === t.id ? 0 : -1"
        @click="selectTab(t.id)"
        @keydown="onTabKey($event, i)">
        {{ t.label }}
      </button>
    </div>

    <div :id="`panel-${tab}`" role="tabpanel" :aria-labelledby="`tab-${tab}`">
      <template v-if="tab === 'controls'">
        <p v-if="status === 'offline'" class="nh-alert is-warn"><UiIcon name="offline" /><span>This device is offline. Values are the last ones it reported and controls are paused.</span></p>
        <p v-else-if="status === 'disabled'" class="nh-alert"><UiIcon name="disabled" /><span>This device is disabled in its settings, so the hub ignores its updates.</span></p>
        <div class="nh-dev-grid">
          <UiCard title="Readings and controls" icon="gauge" flush>
            <ExposeRow v-for="e in measurements" :key="e.name" :device="device" :expose="e" show-description />
            <p v-if="!measurements.length" class="nh-card-body nh-muted">This device reports no readings.</p>
          </UiCard>
          <div class="nh-dev-side">
            <UiCard v-if="configuration.length" title="Configuration" icon="settings" flush>
              <ExposeRow v-for="e in configuration" :key="e.name" :device="device" :expose="e" show-description />
            </UiCard>
            <UiCard v-if="diagnostics.length" title="Diagnostics" icon="signal" flush>
              <ExposeRow v-for="e in diagnostics" :key="e.name" :device="device" :expose="e" />
            </UiCard>
          </div>
        </div>
      </template>
      <DeviceMetrics v-else-if="tab === 'history'" :id="device.id" />
      <UiCard v-else-if="tab === 'settings'" title="Settings on this hub" icon="settings">
        <DeviceSettings :id="device.id" />
      </UiCard>
      <UiCard v-else title="About" icon="info" flush>
        <dl class="nh-kv">
          <dt>Name</dt>
          <dd>{{ device.friendly_name }}</dd>
          <dt>Address</dt>
          <dd><code>{{ device.id }}</code></dd>
          <dt>Model</dt>
          <dd>{{ device.description || '—' }}</dd>
          <dt>Connection</dt>
          <dd>{{ protocol.label }}<span v-if="protocol.generic" class="nh-muted"> (reported as “{{ device.connection_type || 'unknown' }}”)</span></dd>
          <dt>Power</dt>
          <dd>{{ device.power_source || '—' }}</dd>
          <dt>Availability</dt>
          <dd>{{ device.availability }}</dd>
          <dt>Last seen</dt>
          <dd><RelativeTime :value="device.last_seen" /></dd>
        </dl>
      </UiCard>
    </div>
  </div>
  <div v-else-if="!initialized" aria-busy="true" class="nh-card nh-dev-skel"><span class="nh-skel" /><span class="nh-skel" /></div>
  <UiEmpty v-else icon="unknown" title="Device not found" text="It may have been removed or renamed on the bridge.">
    <UiButton :to="{ name: RouteName.DeviceList }" icon="list">Go to the device list</UiButton>
  </UiEmpty>
</template>

<style scoped>
  .nh-dev-hero {
    display: grid;
    place-items: center;
    width: 3.25rem;
    height: 3.25rem;
    border-radius: var(--nh-radius-l);
    background: var(--nh-accent-soft);
    color: var(--nh-accent);
    flex: none;
  }
  .nh-dev-hero .nh-ic {
    width: 1.7rem;
    height: 1.7rem;
  }
  .nh-dev-chips {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.35rem;
    margin: -0.25rem 0 var(--nh-space);
  }
  .nh-dev-seen {
    font-size: 0.8rem;
    margin-left: 0.25rem;
  }
  .nh-tabs {
    display: flex;
    gap: 0.2rem;
    border-bottom: var(--nh-border-w) solid var(--nh-border-strong);
    margin-bottom: var(--nh-space);
    overflow-x: auto;
    scrollbar-width: none;
  }
  .nh-tab {
    padding: 0.65rem 1rem;
    color: var(--nh-text-muted);
    font-weight: 500;
    background: none;
    border: 0;
    border-bottom: 2px solid transparent;
    margin-bottom: -1px;
    white-space: nowrap;
    cursor: pointer;
  }
  .nh-tab[aria-selected='true'] {
    color: var(--nh-accent);
    border-color: var(--nh-accent);
    font-weight: 650;
  }
  .nh-tab:hover {
    color: var(--nh-text);
  }
  .nh-dev-grid {
    display: grid;
    grid-template-columns: minmax(0, 3fr) minmax(0, 2fr);
    gap: var(--nh-space);
    align-items: start;
  }
  .nh-dev-side {
    display: flex;
    flex-direction: column;
    gap: var(--nh-space);
    min-width: 0;
  }
  .nh-dev-side .nh-card + .nh-card {
    margin-top: 0;
  }
  .nh-alert {
    margin: 0 0 var(--nh-space);
  }
  .nh-dev-skel {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    padding: 1rem;
  }
  @media (max-width: 1100px) {
    .nh-dev-grid {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>
