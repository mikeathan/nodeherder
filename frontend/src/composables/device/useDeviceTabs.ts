import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import type { Device } from '@/types/device';
import {
  adjacentDeviceTab,
  deviceTabQuery,
  deviceTabs,
  resolveDeviceTab,
  type DeviceTabId,
} from '@/domain/devices/page';

/**
 * Keeps the device-page tab in the URL and owns keyboard navigation. The available tabs
 * depend on the device, so a link to a withheld tab (for example Presence on a device
 * without mappable coordinates) falls back to Controls instead of showing an empty panel.
 */
export function useDeviceTabs(device: () => Device | undefined) {
  const route = useRoute();
  const router = useRouter();
  const tabs = computed(() => deviceTabs(device()));
  const tab = computed<DeviceTabId>(() => resolveDeviceTab(route.query.tab, tabs.value));

  /** Activates a tab, preserving unrelated query parameters. */
  function selectTab(id: DeviceTabId) {
    router.replace({ query: { ...route.query, tab: deviceTabQuery(id) } });
  }

  /** Moves focus and selection with the left/right arrow keys. */
  function onTabKey(event: KeyboardEvent, index: number) {
    const step = event.key === 'ArrowRight' ? 1 : event.key === 'ArrowLeft' ? -1 : 0;
    if (!step || !tabs.value.length) return;
    const next = adjacentDeviceTab(tabs.value, index, step);
    selectTab(next.id);
    (document.getElementById(`tab-${next.id}`) as HTMLElement | null)?.focus();
  }

  return { tabs, tab, selectTab, onTabKey };
}
