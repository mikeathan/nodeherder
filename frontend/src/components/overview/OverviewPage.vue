<script setup lang="ts">
  /*
   * Overview: one verdict, at most three ranked actions, one line of facts. Network detail
   * moved to Devices and Panel status; full activity stays available but collapsed. The
   * goal is to answer “what first?” in the first viewport.
   */
  import { computed } from 'vue';
  import { store } from '@/store';
  import { RouteName } from '@/types/router';
  import type { Automations } from '@/types/automation.type';
  import type { Device } from '@/types/device';
  import { useHub } from '@/composables/useHub';
  import { useDashboardGroups } from '@/composables/useDashboardGroups';
  import { useAutomationsLoader } from '@/composables/useAutomations';
  import { useNow } from '@/composables/useClock';
  import { summarizeNetwork } from '@/domain/network';
  import { OVERVIEW_REASON_LABEL, overviewIssues } from '@/domain/overview';
  import { relativeTime } from '@/domain/time';
  import { deviceIcon } from '@/components/ui/icons';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import RelativeTime from '@/components/ui/RelativeTime.vue';
  import NotificationDigest from './NotificationDigest.vue';
  import RecentActivityBrief from './RecentActivityBrief.vue';
  import ActivityFeed from './ActivityFeed.vue';

  const { devices, isDeviceDisabled } = useHub();
  const { groups } = useDashboardGroups();
  const automations = useAutomationsLoader();
  const now = useNow();

  const summary = computed(() => summarizeNetwork(devices.value));
  const issues = computed(() => overviewIssues(devices.value, (device: Device) => isDeviceDisabled(device.id), now.value));
  const top = computed(() => issues.value.slice(0, 3));
  const remaining = computed(() => Math.max(0, issues.value.length - top.value.length));
  const enabledRules = computed(() => (automations.value as Automations).filter((automation) => automation.enabled).length);
  const mcp = computed(() => store.getters['hub/mcpStatus']());
  const lastSynced = computed(() => store.getters['hub/lastSyncedAt']() as number | null);
  const snapshot = computed(() =>
    lastSynced.value === null ? 'waiting for the first sync' : `snapshot ${new Date(lastSynced.value).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}`
  );

  /** Short, human breakdown used by the verdict; it never claims a cause the data lacks. */
  const breakdown = computed(() => {
    const counts = new Map<string, number>();
    for (const issue of issues.value) counts.set(issue.reason, (counts.get(issue.reason) ?? 0) + 1);
    return [...counts.entries()].map(([reason, count]) => `${count} ${OVERVIEW_REASON_LABEL[reason as keyof typeof OVERVIEW_REASON_LABEL].toLowerCase()}`).join(' · ');
  });
</script>

<template>
  <div class="nh-overview">
    <UiPageHeader title="Overview" :subtitle="`What needs you, in order · ${snapshot}`">
      <template #actions>
        <UiButton icon="devices" :to="{ name: RouteName.DeviceList }">Devices</UiButton>
      </template>
    </UiPageHeader>

    <NotificationDigest />

    <section class="nh-verdict" :class="{ 'is-clear': !issues.length }" aria-labelledby="nh-verdict-title">
      <span class="nh-verdict-ic"><UiIcon :name="issues.length ? 'warn' : 'check'" /></span>
      <div class="nh-verdict-tx">
        <h2 id="nh-verdict-title">{{ issues.length ? `${issues.length} thing${issues.length === 1 ? '' : 's'} need a look` : 'Everything looks fine' }}</h2>
        <p v-if="issues.length">{{ breakdown }}</p>
        <p v-else>{{ summary.online }} of {{ summary.total }} devices reporting · {{ enabledRules }} automations on · no action needed.</p>
      </div>
      <a v-if="issues.length" class="nh-verdict-cta" href="#nh-start-here">Review in order</a>
    </section>

    <section v-if="top.length" id="nh-start-here" class="nh-start" aria-labelledby="nh-start-title">
      <header>
        <h2 id="nh-start-title">Start here</h2>
        <span>{{ top.length }} of {{ issues.length }}</span>
      </header>
      <ul>
        <li v-for="issue in top" :key="issue.device.id">
          <RouterLink :to="{ name: RouteName.DevicePage, params: { id: issue.device.id } }" class="nh-start-row">
            <span class="nh-start-ic" :class="`is-${issue.reason}`"><UiIcon :path="deviceIcon(issue.device)" /></span>
            <span class="nh-start-tx">
              <b>{{ issue.device.friendly_name }}</b>
              <small>
                {{ OVERVIEW_REASON_LABEL[issue.reason] }}
                <template v-if="issue.device.last_seen"> · last seen <RelativeTime :value="issue.device.last_seen" /></template>
              </small>
            </span>
            <UiIcon name="chevronRight" class="nh-muted" />
          </RouterLink>
        </li>
      </ul>
      <RouterLink v-if="remaining" :to="{ name: RouteName.DeviceList }" class="nh-start-more">
        Show {{ remaining }} more
      </RouterLink>
    </section>

    <div class="nh-health-line">
      <RouterLink :to="{ name: RouteName.DeviceList }">{{ summary.total }} devices · {{ summary.online }} online</RouterLink>
      <span aria-hidden="true">|</span>
      <RouterLink :to="{ name: RouteName.Viewer }">{{ enabledRules }} of {{ automations.length }} automations on</RouterLink>
      <span aria-hidden="true">|</span>
      <RouterLink :to="{ name: RouteName.GroupDashboard }">{{ groups.length }} areas</RouterLink>
      <span aria-hidden="true">|</span>
      <RouterLink :to="{ name: RouteName.Settings }">MCP {{ mcp?.running ? 'on' : 'off' }}</RouterLink>
    </div>

    <section class="nh-recent" aria-labelledby="nh-recent-title">
      <header>
        <h2 id="nh-recent-title">Recent</h2>
        <span>{{ lastSynced === null ? 'no sync yet' : relativeTime(new Date(lastSynced).toISOString(), now) }}</span>
      </header>
      <RecentActivityBrief />
      <details class="nh-activity-details">
        <summary>All activity, filters and recording</summary>
        <ActivityFeed />
      </details>
    </section>
  </div>
</template>

<style scoped>
  .nh-overview {
    max-width: 62rem;
  }
  .nh-verdict {
    display: flex;
    align-items: center;
    gap: 0.9rem;
    padding: 1rem 1.1rem;
    border: var(--nh-border-w) solid color-mix(in srgb, var(--nh-warn) 32%, transparent);
    background: linear-gradient(120deg, var(--nh-warn-soft), transparent);
    border-radius: var(--nh-radius-l);
    margin-bottom: var(--nh-space);
  }
  .nh-verdict.is-clear {
    border-color: color-mix(in srgb, var(--nh-ok) 30%, transparent);
    background: linear-gradient(120deg, var(--nh-ok-soft), transparent);
  }
  .nh-verdict-ic {
    width: 2.6rem;
    height: 2.6rem;
    border-radius: 0.75rem;
    display: grid;
    place-items: center;
    flex: none;
    background: var(--nh-warn-soft);
    color: var(--nh-warn);
  }
  .nh-verdict.is-clear .nh-verdict-ic {
    background: var(--nh-ok-soft);
    color: var(--nh-ok);
  }
  .nh-verdict-tx {
    flex: 1;
    min-width: 0;
  }
  .nh-verdict-tx h2 {
    font-size: 1.1rem;
    font-weight: 700;
  }
  .nh-verdict-tx p {
    margin: 0.15rem 0 0;
    color: var(--nh-text-muted);
    font-size: 0.82rem;
  }
  .nh-verdict-cta {
    flex: none;
    display: inline-flex;
    align-items: center;
    min-height: var(--nh-ctl-h);
    padding: 0 0.9rem;
    border-radius: var(--nh-radius-m);
    background: var(--nh-accent);
    color: var(--nh-accent-contrast);
    font-size: 0.82rem;
    font-weight: 650;
    text-decoration: none;
  }
  .nh-start {
    margin-bottom: var(--nh-space);
  }
  .nh-start header,
  .nh-recent header {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    margin-bottom: 0.4rem;
  }
  .nh-start header h2,
  .nh-recent header h2 {
    font-size: 0.85rem;
    font-weight: 700;
  }
  .nh-start header span,
  .nh-recent header span {
    color: var(--nh-text-muted);
    font-size: 0.72rem;
  }
  .nh-start header span {
    margin-left: auto;
  }
  .nh-start ul {
    list-style: none;
    margin: 0;
    padding: 0;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface);
    overflow: hidden;
  }
  .nh-start li + li {
    border-top: var(--nh-border-w) solid var(--nh-border);
  }
  .nh-start-row {
    display: flex;
    align-items: center;
    gap: 0.7rem;
    padding: 0.65rem 0.75rem;
    color: inherit;
    text-decoration: none;
  }
  .nh-start-row:hover {
    background: var(--nh-surface-2);
  }
  .nh-start-ic {
    width: 2rem;
    height: 2rem;
    border-radius: 0.55rem;
    display: grid;
    place-items: center;
    flex: none;
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
  }
  .nh-start-ic.is-offline,
  .nh-start-ic.is-stale {
    background: var(--nh-danger-soft);
    color: var(--nh-danger);
  }
  .nh-start-ic.is-quiet,
  .nh-start-ic.is-low-battery,
  .nh-start-ic.is-weak-link {
    background: var(--nh-warn-soft);
    color: var(--nh-warn);
  }
  .nh-start-tx {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
  }
  .nh-start-tx b {
    font-size: 0.85rem;
  }
  .nh-start-tx small {
    color: var(--nh-text-muted);
    font-size: 0.75rem;
  }
  .nh-start-more {
    display: inline-block;
    margin-top: 0.45rem;
    font-size: 0.75rem;
    font-weight: 600;
    color: var(--nh-accent);
    text-decoration: none;
  }
  .nh-health-line {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 0.4rem;
    padding: 0.6rem 0.75rem;
    margin-bottom: var(--nh-space);
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
    font-size: 0.78rem;
    color: var(--nh-text-muted);
  }
  .nh-health-line a {
    color: var(--nh-text);
    text-decoration: none;
    border-bottom: 1px dotted var(--nh-border-strong);
  }
  .nh-health-line span {
    color: var(--nh-border-strong);
  }
  .nh-recent {
    margin-bottom: var(--nh-space);
  }
  .nh-activity-details {
    margin-top: 0.7rem;
  }
  .nh-activity-details summary {
    cursor: pointer;
    color: var(--nh-text-muted);
    font-size: 0.75rem;
  }
  @media (max-width: 640px) {
    .nh-verdict {
      align-items: flex-start;
      flex-wrap: wrap;
    }
    .nh-verdict-cta {
      margin-left: 3.5rem;
    }
  }
</style>
