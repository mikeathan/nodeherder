<script setup lang="ts">
  /*
   * Panel mode (spec 007 US-07, FR-10, AC-23/24): a full-screen view for a wall tablet with a
   * clock, one Home area at a time (swipe-free pager, arrow keys) and a network status page.
   * Uses the same commands, themes and data as the rest of the app.
   */
  import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
  import { RouteName } from '@/types/router';
  import { Automations } from '@/types/automation.type';
  import { useHomeLayout } from '@/composables/useHomeLayout';
  import { homeDeviceGroup } from '@/domain/home-compatibility';
  import { homePanelBlocks, homePanelItems } from '@/domain/home-panel';
  import { homeCardTitle } from '@/domain/home-cards';
  import IntegrationTile from '@/components/dashboards/IntegrationTile.vue';
  import { findDevice, useHub } from '@/composables/useHub';
  import { useAutomationsLoader } from '@/composables/useAutomations';
  import { useThemeSettings } from '@/composables/useThemeSettings';
  import { useNow } from '@/composables/useClock';
  import { normalizeRoomIndex, panelRoom } from '@/domain/panel';
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
  import DefaultViewButton from '@/components/ui/DefaultViewButton.vue';
  import PanelDial from './PanelDial.vue';
  import PanelButton from './PanelButton.vue';

  defineEmits<{ (e: 'open-appearance'): void }>();

  const ROOM_KEY = 'nodeherder_panel_room';
  const home = useHomeLayout();
  const { sources, sourceError } = home;
  const selectedId = ref<string | null>(null);
  const groups = computed(() => homePanelBlocks(home.layout.value));
  const { devices, isDeviceDisabled } = useHub();
  const automations = useAutomationsLoader();
  const { mode, toggleMode } = useThemeSettings();
  const now = useNow(15_000);

  const page = ref<'rooms' | 'status'>('rooms');
  const roomIndex = ref(Number(safeStorage.getItem(ROOM_KEY)) || 0);
  watch(roomIndex, (i) => safeStorage.setItem(ROOM_KEY, String(i)));

  const roomTabs = ref<HTMLElement | null>(null);
  const current = computed(() => {
    const index = selectedId.value ? groups.value.findIndex((block) => block.id === selectedId.value) : -1;
    return index >= 0 ? index : normalizeRoomIndex(roomIndex.value, groups.value.length);
  });
  const block = computed(() => groups.value[current.value]);
  const room = computed(() => (block.value ? panelRoom(homeDeviceGroup(block.value), findDevice) : null));
  const panelItems = computed(() => (block.value && room.value ? homePanelItems(block.value, room.value) : []));
  function pageName(index: number): string {
    const group = groups.value[index];
    if (group.kind === 'group') return group.name;
    const card = group.cards[0];
    return homeCardTitle(card, sources.value);
  }
  function go(index: number) {
    roomIndex.value = normalizeRoomIndex(index, groups.value.length);
    selectedId.value = groups.value[roomIndex.value]?.id ?? null;
  }

  const time = computed(() => new Date(now.value).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }));
  const date = computed(() =>
    new Date(now.value).toLocaleDateString([], { weekday: 'long', day: 'numeric', month: 'long' })
  );

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
  const avgLqi = computed(() =>
    lqis.value.length ? Math.round(lqis.value.reduce((a, b) => a + b, 0) / lqis.value.length) : 0
  );
  const batteries = computed(() => devices.value.filter((d) => isBatteryPowered(d) && batteryLevel(d) !== null));
  const healthyBatteries = computed(() => batteries.value.filter((d) => !hasLowBattery(d)).length);
  const rulesOn = computed(() => (automations.value as Automations).filter((a) => a.enabled).length);
  const attention = computed(() => needsAttention(devices.value, (d) => isDeviceDisabled(d.id)).slice(0, 8));

  function modifiedKey(e: KeyboardEvent) {
    return e.defaultPrevented || e.altKey || e.ctrlKey || e.metaKey || e.shiftKey;
  }
  function onRoomKey(e: KeyboardEvent) {
    if (modifiedKey(e)) return;
    const destinations: Record<string, number> = {
      ArrowLeft: current.value - 1,
      ArrowRight: current.value + 1,
      Home: 0,
      End: groups.value.length - 1,
    };
    if (!Object.prototype.hasOwnProperty.call(destinations, e.key)) return;
    e.preventDefault();
    go(destinations[e.key]);
    void nextTick(() => roomTabs.value?.querySelectorAll<HTMLButtonElement>('[role="tab"]')[current.value]?.focus());
  }
  function onKey(e: KeyboardEvent) {
    if (modifiedKey(e) || page.value !== 'rooms' || !/^Arrow(Left|Right)$/.test(e.key)) return;
    const target = e.target instanceof HTMLElement ? e.target : null;
    if (
      target?.isContentEditable ||
      target?.closest('input, select, textarea, button, a, [role="dialog"], [role="slider"], [role="spinbutton"]')
    )
      return;
    e.preventDefault();
    go(current.value + (e.key === 'ArrowLeft' ? -1 : 1));
  }
  watch([current, page], () => {
    // Arrow buttons can select a room outside the horizontal pager viewport.
    void nextTick(() =>
      roomTabs.value
        ?.querySelector<HTMLElement>('[aria-selected="true"]')
        ?.scrollIntoView({ block: 'nearest', inline: 'nearest' })
    );
  });
  onMounted(() => window.addEventListener('keydown', onKey));
  onBeforeUnmount(() => window.removeEventListener('keydown', onKey));
</script>

<template>
  <div class="hp">
    <header class="hp-strip">
      <RouterLink :to="{ name: RouteName.GroupDashboard }" class="hp-logo" aria-label="Exit panel mode">
        <AppLogo :size="46" :wordmark="false" />
      </RouterLink>
      <div class="hp-clock">
        <b>{{ time }}</b>
        <span>{{ date }}</span>
      </div>
      <span class="hp-gap" />
      <ConnectionStatus />
      <UiButton
        variant="ghost"
        :icon="mode === 'dark' ? 'light' : 'dark'"
        :label="mode === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
        @click="toggleMode" />
      <UiButton variant="ghost" icon="palette" label="Appearance" @click="$emit('open-appearance')" />
    </header>
    <div class="hp-notices">
      <StatusBanners />
      <p v-if="home.error.value" class="nh-alert is-warning" role="alert">{{ home.error.value }}</p>
      <p v-if="sourceError" class="nh-alert is-warning" role="status">
        Integration results could not be updated. Device controls remain available.
      </p>
    </div>

    <main id="nh-content" class="hp-content" tabindex="-1">
      <template v-if="page === 'rooms'">
        <UiEmpty
          v-if="!room"
          icon="home"
          title="No rooms yet"
          text="Panel mode shows your Home areas. Create one on the Home page.">
          <UiButton :to="{ name: RouteName.GroupDashboard }">Go to Home</UiButton>
        </UiEmpty>
        <template v-else>
          <div class="hp-pager">
            <UiButton icon="chevronLeft" label="Previous room" @click="go(current - 1)" />
            <div ref="roomTabs" class="hp-pager-list" role="tablist" aria-label="Rooms" @keydown="onRoomKey">
              <button
                v-for="(g, i) in groups"
                :key="g.id"
                type="button"
                role="tab"
                :aria-selected="i === current"
                :tabindex="i === current ? 0 : -1"
                :id="`panel-room-tab-${i}`"
                aria-controls="panel-room"
                @click="go(i)">
                {{ pageName(i) }}
              </button>
            </div>
            <UiButton icon="chevronRight" label="Next room" @click="go(current + 1)" />
          </div>
          <section
            class="hp-room"
            :class="{ 'no-climate': !climate }"
            id="panel-room"
            :aria-labelledby="`panel-room-tab-${current}`"
            role="tabpanel">
            <div v-if="climate" class="hp-climate">
              <PanelDial
                v-if="climate.online"
                :value="climate.value"
                :min="5"
                :max="35"
                :label="climate.label"
                :caption="climate.caption" />
              <div v-else class="hp-dial-off">
                <UiIcon name="offline" />
                <b>Offline</b>
                <span>{{ climate.name }}</span>
              </div>
            </div>
            <div class="hp-buttons">
              <template v-for="item in panelItems" :key="item.id">
                <PanelButton
                  v-if="item.kind === 'device'"
                  :device-id="item.button.deviceId"
                  :expose="item.button.expose"
                  :label="item.button.label" />
                <IntegrationTile v-else :card="item.card" :sources="sources" :source-error="sourceError" panel />
              </template>
              <p v-if="!panelItems.length && !climate" class="nh-muted">This area has no tiles yet.</p>
            </div>
          </section>
        </template>
      </template>

      <template v-else>
        <div class="hp-gauges">
          <PanelDial
            :value="online"
            :min="0"
            :max="devices.length"
            :label="`${online}/${devices.length}`"
            caption="devices online"
            kind="ok" />
          <PanelDial
            :value="avgLqi"
            :min="0"
            :max="255"
            :label="lqis.length ? String(avgLqi) : '—'"
            caption="average link quality"
            kind="signal" />
          <PanelDial
            :value="healthyBatteries"
            :min="0"
            :max="batteries.length"
            :label="`${healthyBatteries}/${batteries.length}`"
            caption="batteries healthy"
            kind="lamp" />
          <PanelDial
            :value="rulesOn"
            :min="0"
            :max="automations.length"
            :label="`${rulesOn}/${automations.length}`"
            caption="automations on"
            kind="lamp" />
        </div>
        <h2 class="hp-h2">Needs a look</h2>
        <div class="hp-rows">
          <RouterLink
            v-for="a in attention"
            :key="a.device.id"
            :to="{ name: RouteName.DevicePage, params: { id: a.device.id } }"
            class="hp-row">
            <UiIcon :path="deviceIcon(a.device)" />
            <b>{{ a.device.friendly_name }}</b>
            <span>
              {{
                a.reason === 'offline'
                  ? 'Offline'
                  : a.reason === 'low-battery'
                    ? 'Low battery'
                    : a.reason === 'weak-link'
                      ? 'Weak signal'
                      : 'Disabled'
              }}
            </span>
          </RouterLink>
          <p v-if="!attention.length" class="hp-row">
            <UiIcon name="check" />
            <b>Everything looks fine</b>
          </p>
        </div>
      </template>
    </main>

    <nav class="hp-dock" aria-label="Panel">
      <DefaultViewButton dock />
      <button type="button" :aria-pressed="page === 'rooms'" @click="page = 'rooms'">
        <UiIcon name="home" />
        <span>Rooms</span>
      </button>
      <button type="button" :aria-pressed="page === 'status'" @click="page = 'status'">
        <UiIcon name="gauge" />
        <span>Status</span>
      </button>
      <RouterLink :to="{ name: RouteName.GroupDashboard }">
        <UiIcon name="signOut" />
        <span>Exit panel</span>
      </RouterLink>
    </nav>
  </div>
</template>

<style scoped>
  .hp {
    --hp-dock-target: 2.75rem;
    --hp-dock-padding: 0.3rem;
    --hp-dock-bottom-gap: max(0.6rem, env(safe-area-inset-bottom));
    /* Share dock geometry with notices: no overlap, layout shift or resize observer. */
    --nh-notice-bottom: calc(
      var(--hp-dock-target) + var(--hp-dock-padding) * 2 + var(--nh-border-w) * 2 + var(--hp-dock-bottom-gap) + 0.5rem
    );
    --nh-notice-max-height: calc(100dvh - var(--nh-header-h) - var(--nh-notice-bottom) - 1rem);
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
    white-space: nowrap;
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
  /* Keep the dock in its grid row: compact controls never cover scrollable content. */
  .hp-dock {
    display: flex;
    justify-content: center;
    gap: 0.25rem;
    padding: var(--hp-dock-padding);
    margin: 0 auto var(--hp-dock-bottom-gap);
    max-width: calc(100% - 1rem);
    background: var(--nh-surface);
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: 18px;
  }
  .hp-dock button,
  .hp-dock a {
    display: flex;
    flex-direction: row;
    align-items: center;
    justify-content: center;
    gap: 0.4rem;
    min-width: var(--hp-dock-target);
    min-height: var(--hp-dock-target);
    padding: 0.4rem 0.7rem;
    border-radius: 12px;
    border: 0;
    background: none;
    color: var(--nh-text-muted);
    font: inherit;
    font-size: 0.8rem;
    white-space: nowrap;
    text-decoration: none;
    cursor: pointer;
  }
  .hp-dock :deep(.nh-ic) {
    width: 1.15rem;
    height: 1.15rem;
  }
  .hp-dock button:hover,
  .hp-dock a:hover {
    background: var(--nh-surface-2);
    color: var(--nh-text);
  }
  .hp-dock button[aria-pressed='true']:not(.nh-default-view) {
    background: var(--nh-accent);
    color: var(--nh-accent-contrast);
  }
  /* Pin is a saved preference, distinct from the current Rooms/Status selection. */
  .hp-dock button.nh-default-view[aria-pressed='true'] {
    background: var(--nh-surface-2);
    color: var(--nh-accent);
    box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--nh-accent) 40%, transparent);
  }
  @media (max-width: 760px) {
    .hp {
      --hp-dock-padding: 0.25rem;
      --hp-dock-bottom-gap: max(0.5rem, env(safe-area-inset-bottom));
    }
    .hp-strip {
      gap: 0.4rem;
      padding: 0.6rem 1rem;
    }
    .hp-clock b {
      font-size: clamp(1.25rem, 6vw, 2rem);
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
    .hp-dock button,
    .hp-dock a {
      gap: 0.3rem;
      padding: 0.35rem 0.5rem;
      font-size: 0.75rem;
    }
  }
</style>
