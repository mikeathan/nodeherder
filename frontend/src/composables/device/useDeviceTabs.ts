import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { DEVICE_TABS, adjacentDeviceTab, deviceTabQuery, resolveDeviceTab, type DeviceTabId } from '@/domain/devices/page';

/** Keeps the device-page tab in the URL and owns keyboard navigation. */
export function useDeviceTabs() {
  const route = useRoute();
  const router = useRouter();
  const tab = computed<DeviceTabId>(() => resolveDeviceTab(route.query.tab));

  /** Activates a tab, preserving unrelated query parameters. */
  function selectTab(id: DeviceTabId) {
    router.replace({ query: { ...route.query, tab: deviceTabQuery(id) } });
  }

  /** Moves focus and selection with the left/right arrow keys. */
  function onTabKey(event: KeyboardEvent, index: number) {
    const step = event.key === 'ArrowRight' ? 1 : event.key === 'ArrowLeft' ? -1 : 0;
    if (!step) return;
    const next = adjacentDeviceTab(index, step);
    selectTab(next.id);
    (document.getElementById(`tab-${next.id}`) as HTMLElement | null)?.focus();
  }

  return { tabs: DEVICE_TABS, tab, selectTab, onTabKey };
}
