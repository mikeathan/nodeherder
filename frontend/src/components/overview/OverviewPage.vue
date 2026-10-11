<script setup lang="ts">
  /* Overview (spec 007 US-01, FR-05): network health, what needs a look, recent activity. */
  import { computed } from 'vue';
  import { store } from '@/store';
  import { RouteName } from '@/types/router';
  import { Automations } from '@/types/automation.type';
  import { useHub } from '@/composables/useHub';
  import { useDashboardGroups } from '@/composables/useDashboardGroups';
  import { useAutomationsLoader } from '@/composables/useAutomations';
  import { needsAttention, summarizeNetwork } from '@/domain/network';
  import { batteryLevel, linkQuality } from '@/domain/devices';
  import { isKnownTime } from '@/domain/time';
  import { deviceIcon } from '@/components/ui/icons';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import RelativeTime from '@/components/ui/RelativeTime.vue';
  import StatTile from './StatTile.vue';
  import ActivityFeed from './ActivityFeed.vue';

  const { devices, isDeviceDisabled } = useHub();
  const { groups } = useDashboardGroups();
  const automations = useAutomationsLoader();

  const summary = computed(() => summarizeNetwork(devices.value));
  const attention = computed(() => needsAttention(devices.value, (d) => isDeviceDisabled(d.id)));
  const enabledRules = computed(() => (automations.value as Automations).filter((a) => a.enabled).length);
  const mcp = computed(() => store.getters['hub/mcpStatus']());

  const reasonText = {
    offline: 'Offline',
    disabled: 'Disabled',
    'low-battery': 'Low battery',
    'weak-link': 'Weak signal',
  } as const;
</script>

<template>
  <div>
    <UiPageHeader title="Overview" subtitle="How your home network is doing right now." />

    <div class="nh-stats">
      <StatTile :to="{ name: RouteName.DeviceList }" icon="devices" :value="`${summary.online}/${summary.total}`" label="Devices online" :tone="summary.offline ? 'warn' : 'ok'" :hint="summary.offline ? `${summary.offline} offline` : 'All reachable'" />
      <StatTile :to="{ name: RouteName.DeviceList }" icon="warn" :value="String(attention.length)" label="Need a look" :tone="attention.length ? 'warn' : 'ok'" hint="Offline, low battery, weak signal" />
      <StatTile :to="{ name: RouteName.Viewer }" icon="automation" :value="`${enabledRules}/${automations.length}`" label="Automations on" />
      <StatTile :to="{ name: RouteName.GroupDashboard }" icon="home" :value="String(groups.length)" label="Home areas" />
      <StatTile :to="{ name: RouteName.Settings }" icon="assistant" :value="mcp?.running ? 'On' : 'Off'" label="MCP server" :hint="mcp?.running ? `${mcp.connectedClients ?? 0} client(s)` : 'Assistant tools'" />
    </div>

    <div class="nh-cols">
      <ActivityFeed />
      <div class="nh-side">
        <UiCard title="Needs a look" icon="warn" flush>
          <ul class="nh-attn">
            <li v-for="item in attention.slice(0, 8)" :key="item.device.id">
              <RouterLink :to="{ name: RouteName.DevicePage, params: { id: item.device.id } }">
                <span class="nh-attn-ic" :class="`is-${item.reason}`"><UiIcon :path="deviceIcon(item.device)" /></span>
                <span class="nh-attn-t">
                  <b>{{ item.device.friendly_name }}</b>
                  <small>
                    {{ reasonText[item.reason] }}
                    <template v-if="item.reason === 'offline' && isKnownTime(item.device.last_seen)"> · last seen <RelativeTime :value="item.device.last_seen" /></template>
                    <template v-else-if="item.reason === 'low-battery'"> · {{ Math.round(batteryLevel(item.device) ?? 0) }} %</template>
                    <template v-else-if="item.reason === 'weak-link'"> · {{ linkQuality(item.device) }} LQI</template>
                  </small>
                </span>
                <UiIcon name="chevronRight" class="nh-muted" />
              </RouterLink>
            </li>
            <li v-if="!attention.length" class="nh-attn-ok"><UiIcon name="check" />Everything looks fine.</li>
          </ul>
        </UiCard>
        <UiCard title="Network" icon="zigbee" flush>
          <dl class="nh-kv">
            <dt>Zigbee routers</dt>
            <dd>{{ summary.zigbeeRouters }}</dd>
            <dt>Zigbee end devices</dt>
            <dd>{{ summary.zigbeeEndDevices }}</dd>
            <dt>Wi-Fi and other</dt>
            <dd>{{ summary.otherProtocols }}</dd>
            <dt>Weak signal</dt>
            <dd>{{ summary.weakLink }}</dd>
            <dt>Low battery</dt>
            <dd>{{ summary.lowBattery }}</dd>
          </dl>
        </UiCard>
      </div>
    </div>
  </div>
</template>

<style scoped>
  .nh-stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 10rem), 1fr));
    gap: var(--nh-gap);
    margin-bottom: calc(var(--nh-space) * 1.2);
  }
  .nh-cols {
    display: grid;
    grid-template-columns: minmax(0, 2fr) minmax(0, 1fr);
    gap: var(--nh-space);
    align-items: start;
  }
  .nh-side {
    display: flex;
    flex-direction: column;
    gap: var(--nh-space);
    min-width: 0;
  }
  .nh-side .nh-card + .nh-card {
    margin-top: 0;
  }
  .nh-attn {
    list-style: none;
    margin: 0;
    padding: 0.3rem 0;
  }
  .nh-attn a {
    display: flex;
    align-items: center;
    gap: 0.7rem;
    padding: 0.5rem var(--nh-space);
    text-decoration: none;
  }
  .nh-attn a:hover {
    background: var(--nh-surface-2);
  }
  .nh-attn-t {
    flex: 1;
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .nh-attn-t b {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .nh-attn-t small {
    color: var(--nh-text-muted);
  }
  .nh-attn-ic {
    display: grid;
    color: var(--nh-warn);
  }
  .nh-attn-ic.is-offline {
    color: var(--nh-danger);
  }
  .nh-attn-ic.is-disabled {
    color: var(--nh-text-muted);
  }
  .nh-attn-ok {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.6rem var(--nh-space);
    color: var(--nh-text-muted);
  }
  .nh-attn-ok .nh-ic {
    color: var(--nh-ok);
  }
  @media (max-width: 1100px) {
    .nh-cols {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>
