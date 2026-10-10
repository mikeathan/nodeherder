<script setup lang="ts">
  /*
   * Panel mode (spec 007 US-07, FR-10, AC-23/24): a full-screen view for a wall tablet with a
   * clock, one Home area at a time (swipe-free pager, arrow keys) and a network status page.
   * Uses the same commands, themes and data as the rest of the app.
   */
  import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
  import { RouteName } from '@/types/router';
  import { Automations } from '@/types/automation.type';
  import { useDashboardGroups } from '@/composables/useDashboardGroups';
  import { findDevice, useHub } from '@/composables/useHub';
  import { useAutomationsLoader } from '@/composables/useAutomations';
  import { useThemeSettings } from '@/composables/useThemeSettings';
  import { useNow } from '@/composables/useClock';
  import { panelRoom } from '@/domain/panel';
  import { formatExposeValue } from '@/domain/exposes';
  import { isOnline, isBatteryPowered, hasLowBattery, linkQuality, batteryLevel } from '@/domain/devices';
  import { needsAttention } from '@/domain/network';
  import { deviceIcon } from '@/components/ui/icons';
  import { safeStorage } from '@/utils/storage';
  import AppLogo from '@/components/ui/AppLogo.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import UiEmpty from '@/components/ui/UiEmpty.vue';
  import ConnectionStatus from '@/components/shell/ConnectionStatus.vue';
  import StatusBanners from '@/components/shell/StatusBanners.vue';
  import PanelDial from './PanelDial.vue';
  import PanelButton from './PanelButton.vue';

  defineEmits<{ (e: 'open-appearance'): void }>();

  const ROOM_KEY = 'nodeherder_panel_room';
  const { groups } = useDashboardGroups();
  const { devices, isDeviceDisabled } = useHub();
  const automations = useAutomationsLoader();
  const { mode, toggleMode } = useThemeSettings();
  const now = useNow(15_000);

  const page = ref<'rooms' | 'status'>('rooms');
  const roomIndex = ref(Number(safeStorage.getItem(ROOM_KEY)) || 0);
  watch(roomIndex, (i) => safeStorage.setItem(ROOM_KEY, String(i)));

  const current = computed(() => {
    const n = groups.value.length;
    return n ? ((roomIndex.value % n) + n) % n : 0;
  });
  const room = computed(() => (groups.value.length ? panelRoom(groups.value[current.value], findDevice) : null));
  const go = (i: number) => (roomIndex.value = i);

  const time = computed(() => new Date(now.value).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }));
  const date = computed(() => new Date(now.value).toLocaleDateString([], { weekday: 'long', day: 'numeric', month: 'long' }));

  const climate = computed(() => {
    const c = room.value?.climate;
    if (!c) return null;
    const d = findDevice(c.deviceId);
    const t = d?.exposes[c.expose];
    if (!d || !t) return null;
    const h = room.value?.humidity;
    const hd = h ? findDevice(h.deviceId) : undefined;
    const hum = h && hd && isOnline(hd) ? hd.exposes[h.expose] : undefined;
    return {
      online: isOnline(d) && typeof t.data === 'number',
      value: Number(t.data),
      label: formatExposeValue(t),
      caption: hum ? `${formatExposeValue(hum)} humidity` : 'temperature',
      name: d.friendly_name,
    };
  });

  // status page
  const online = computed(() => devices.value.filter(isOnline).length);
  const lqis = computed(() => devices.value.map(linkQuality).filter((v): v is number => v !== null));
  const avgLqi = computed(() => (lqis.value.length ? Math.round(lqis.value.reduce((a, b) => a + b, 0) / lqis.value.length) : 0));
  const batteries = computed(() => devices.value.filter((d) => isBatteryPowered(d) && batteryLevel(d) !== null));
  const healthyBatteries = computed(() => batteries.value.filter((d) => !hasLowBattery(d)).length);
  const rulesOn = computed(() => (automations.value as Automations).filter((a) => a.enabled).length);
  const attention = computed(() => needsAttention(devices.value, (d) => isDeviceDisabled(d.id)).slice(0, 8));

  function onKey(e: KeyboardEvent) {
    if (page.value !== 'rooms' || !/^Arrow(Left|Right)$/.test(e.key)) return;
    if (/INPUT|SELECT|TEXTAREA/.test((document.activeElement as HTMLElement | null)?.tagName ?? '')) return;
    go(current.value + (e.key === 'ArrowLeft' ? -1 : 1));
  }
  onMounted(() => window.addEventListener('keydown', onKey));
  onBeforeUnmount(() => window.removeEventListener('keydown', onKey));
</script>

<template>
  <div class="hp">
    <header class="hp-strip">
      <RouterLink :to="{ name: RouteName.GroupDashboard }" class="hp-logo" aria-label="Exit panel mode"><AppLogo :size="46" :wordmark="false" /></RouterLink>
      <div class="hp-clock">
        <b>{{ time }}</b>
        <span>{{ date }}</span>
      </div>
      <span class="hp-gap" />
      <ConnectionStatus />
      <UiButton variant="ghost" :icon="mode === 'dark' ? 'light' : 'dark'" :label="mode === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'" @click="toggleMode" />
      <UiButton variant="ghost" icon="palette" label="Appearance" @click="$emit('open-appearance')" />
    </header>
    <StatusBanners />

    <main id="nh-content" class="hp-content" tabindex="-1">
      <template v-if="page === 'rooms'">
        <UiEmpty v-if="!room" icon="home" title="No rooms yet" text="Panel mode shows your Home areas. Create one on the Home page.">
          <UiButton :to="{ name: RouteName.GroupDashboard }">Go to Home</UiButton>
        </UiEmpty>
        <template v-else>
          <div class="hp-pager">
            <UiButton icon="chevronLeft" label="Previous room" @click="go(current - 1)" />
            <div class="hp-pager-list" role="tablist" aria-label="Rooms">
              <button
                v-for="(g, i) in groups"
                :key="g.name"
                type="button"
                role="tab"
                :aria-selected="i === current"
                @click="go(i)">
                {{ g.name }}
              </button>
            </div>
            <UiButton icon="chevronRight" label="Next room" @click="go(current + 1)" />
          </div>
          <section class="hp-room" :class="{ 'no-climate': !climate }" :aria-label="room.name" role="tabpanel">
            <div v-if="climate" class="hp-climate">
              <PanelDial v-if="climate.online" :value="climate.value" :min="5" :max="35" :label="climate.label" :caption="climate.caption" />
              <div v-else class="hp-dial-off"><UiIcon name="offline" /><b>Offline</b><span>{{ climate.name }}</span></div>
            </div>
            <div class="hp-buttons">
              <PanelButton v-for="b in room.buttons" :key="`${b.deviceId}|${b.expose}`" :device-id="b.deviceId" :expose="b.expose" :label="b.label" />
              <p v-if="!room.buttons.length && !climate" class="nh-muted">This area has no tiles yet.</p>
            </div>
          </section>
        </template>
      </template>

      <template v-else>
        <div class="hp-gauges">
          <PanelDial :value="online" :min="0" :max="devices.length" :label="`${online}/${devices.length}`" caption="devices online" kind="ok" />
          <PanelDial :value="avgLqi" :min="0" :max="255" :label="lqis.length ? String(avgLqi) : '—'" caption="average link quality" kind="signal" />
          <PanelDial :value="healthyBatteries" :min="0" :max="batteries.length" :label="`${healthyBatteries}/${batteries.length}`" caption="batteries healthy" kind="lamp" />
          <PanelDial :value="rulesOn" :min="0" :max="automations.length" :label="`${rulesOn}/${automations.length}`" caption="automations on" kind="lamp" />
        </div>
        <h2 class="hp-h2">Needs a look</h2>
        <div class="hp-rows">
          <RouterLink v-for="a in attention" :key="a.device.id" :to="{ name: RouteName.DevicePage, params: { id: a.device.id } }" class="hp-row">
            <UiIcon :path="deviceIcon(a.device)" />
            <b>{{ a.device.friendly_name }}</b>
            <span>{{ a.reason === 'offline' ? 'Offline' : a.reason === 'low-battery' ? 'Low battery' : a.reason === 'weak-link' ? 'Weak signal' : 'Disabled' }}</span>
          </RouterLink>
          <p v-if="!attention.length" class="hp-row"><UiIcon name="check" /><b>Everything looks fine</b></p>
        </div>
      </template>
    </main>

    <nav class="hp-dock" aria-label="Panel">
      <button type="button" :aria-pressed="page === 'rooms'" @click="page = 'rooms'"><UiIcon name="home" /><span>Rooms</span></button>
      <button type="button" :aria-pressed="page === 'status'" @click="page = 'status'"><UiIcon name="gauge" /><span>Status</span></button>
      <RouterLink :to="{ name: RouteName.GroupDashboard }"><UiIcon name="signOut" /><span>Exit panel</span></RouterLink>
    </nav>
  </div>
</template>

<style scoped>
  .hp {
    display: grid;
    grid-template-columns: minmax(0, 1fr);
    grid-template-rows: auto auto minmax(0, 1fr) auto;
    height: 100dvh;
    font-family: var(--nh-font-panel);
    font-weight: 300;
    background:
      radial-gradient(60rem 40rem at 10% -10%, color-mix(in srgb, var(--dk-climate) 16%, transparent), transparent 70%),
      radial-gradient(50rem 40rem at 110% 110%, color-mix(in srgb, var(--dk-lamp) 12%, transparent), transparent 70%),
      var(--nh-bg);
  }
  .hp :deep(b),
  .hp h2 {
    font-weight: 500;
  }
  .hp-strip {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.8rem 1.5rem;
  }
  .hp-logo {
    display: block;
    flex: none;
  }
  .hp-clock {
    display: flex;
    align-items: baseline;
    gap: 0.8rem;
    min-width: 0;
  }
  .hp-clock b {
    font-size: 2.6rem;
    font-weight: 300 !important;
    font-variant-numeric: tabular-nums;
    letter-spacing: -0.02em;
  }
  .hp-clock span {
    color: var(--nh-text-muted);
    white-space: nowrap;
  }
  .hp-gap {
    flex: 1;
  }
  .hp-strip :deep(.nh-btn.is-icon) {
    width: 2.6rem;
    height: 2.6rem;
    border-radius: 50%;
    background: var(--dk-glass);
  }
  .hp-content {
    overflow: auto;
    padding: 0.5rem 1.5rem 1.5rem;
    outline: none;
  }
  .hp-pager {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin-bottom: 1.25rem;
  }
  .hp-pager > :deep(.nh-btn.is-icon) {
    width: 3rem;
    height: 3rem;
    border-radius: 50%;
    flex: none;
  }
  .hp-pager-list {
    display: flex;
    gap: 0.2rem;
    overflow-x: auto;
    scroll-snap-type: x proximity;
    flex: 1;
    min-width: 0;
    scrollbar-width: none;
  }
  .hp-pager-list button {
    border: 0;
    background: none;
    color: var(--nh-text-muted);
    font: inherit;
    font-size: 1.25rem;
    padding: 0.5rem 1rem;
    border-radius: 99px;
    cursor: pointer;
    white-space: nowrap;
    scroll-snap-align: start;
  }
  .hp-pager-list button[aria-selected='true'] {
    color: var(--nh-text);
    background: var(--nh-surface-2);
    font-weight: 500;
  }
  .hp-room {
    display: grid;
    grid-template-columns: minmax(15rem, 22rem) minmax(0, 1fr);
    gap: 1.5rem;
    align-items: start;
  }
  .hp-room.no-climate {
    grid-template-columns: minmax(0, 1fr);
  }
  .hp-climate {
    background: var(--nh-surface);
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    padding: 1.25rem;
  }
  .hp-dial-off {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.4rem;
    padding: 3rem 1rem;
    color: var(--nh-text-muted);
  }
  .hp-dial-off .nh-ic {
    width: 2.5rem;
    height: 2.5rem;
    color: var(--nh-danger);
  }
  .hp-buttons {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(10.5rem, 1fr));
    gap: 1rem;
  }
  .hp-gauges {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
    gap: 1rem;
    margin-bottom: 1.5rem;
  }
  .hp-gauges > * {
    background: var(--nh-surface);
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    padding: 1rem;
  }
  .hp-gauges :deep(figcaption b) {
    font-size: 2rem;
  }
  .hp-h2 {
    font-size: 1.3rem;
    margin: 0 0 0.75rem;
  }
  .hp-rows {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  .hp-row {
    display: flex;
    align-items: center;
    gap: 1rem;
    min-height: 4rem;
    padding: 0.6rem 1.1rem;
    margin: 0;
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface);
    border: var(--nh-border-w) solid var(--nh-border);
    text-decoration: none;
  }
  .hp-row > .nh-ic {
    width: 1.7rem;
    height: 1.7rem;
    color: var(--nh-accent);
  }
  .hp-row span {
    color: var(--nh-text-muted);
    margin-left: auto;
    text-align: right;
  }
  .hp-dock {
    display: flex;
    justify-content: center;
    gap: 0.4rem;
    padding: 0.5rem;
    margin: 0 auto 0.9rem;
    background: var(--nh-surface);
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: 26px;
  }
  .hp-dock button,
  .hp-dock a {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.2rem;
    min-width: 5.5rem;
    min-height: 4rem;
    padding: 0.4rem 0.8rem;
    border-radius: 18px;
    border: 0;
    background: none;
    color: var(--nh-text-muted);
    font: inherit;
    font-size: 0.85rem;
    text-decoration: none;
    cursor: pointer;
  }
  .hp-dock .nh-ic {
    width: 1.6rem;
    height: 1.6rem;
  }
  .hp-dock button:hover,
  .hp-dock a:hover {
    background: var(--nh-surface-2);
    color: var(--nh-text);
  }
  .hp-dock [aria-pressed='true'] {
    background: var(--nh-accent);
    color: var(--nh-accent-contrast);
  }
  @media (max-width: 760px) {
    .hp-strip {
      gap: 0.4rem;
      padding: 0.6rem 1rem;
    }
    .hp-clock b {
      font-size: 2rem;
    }
    .hp-clock span {
      display: none;
    }
    .hp-content {
      padding: 0.25rem 1rem 1rem;
    }
    .hp-room {
      grid-template-columns: minmax(0, 1fr);
      gap: 1rem;
    }
    .hp-climate {
      padding: 0.75rem;
    }
    .hp-climate :deep(svg) {
      max-width: 12rem;
    }
    .hp-buttons {
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 0.75rem;
    }
    .hp-pager-list button {
      font-size: 1.05rem;
      padding: 0.45rem 0.8rem;
    }
    .hp-pager > :deep(.nh-btn.is-icon) {
      width: 2.6rem;
      height: 2.6rem;
    }
    .hp-gauges {
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 0.75rem;
    }
    .hp-gauges :deep(figcaption b) {
      font-size: 1.5rem;
    }
    .hp-gauges :deep(figcaption span) {
      font-size: 0.8rem;
    }
    .hp-row {
      flex-wrap: wrap;
    }
    .hp-row span {
      flex-basis: 100%;
      margin-left: 2.7rem;
      text-align: left;
    }
    .hp-dock {
      margin: 0;
      border-radius: 0;
      border-width: 1px 0 0;
      padding-bottom: max(0.5rem, env(safe-area-inset-bottom));
    }
    .hp-dock button,
    .hp-dock a {
      flex: 1;
      min-width: 0;
    }
  }
</style>
