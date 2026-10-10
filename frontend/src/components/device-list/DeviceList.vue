<script setup lang="ts">
  /* Device list (spec 007 US-03, AC-07…09): search, filters, sortable columns, bridge actions. */
  import { computed, ref } from 'vue';
  import { RouteName } from '@/types/router';
  import { useHub } from '@/composables/useHub';
  import { DEFAULT_FILTERS, DeviceFilters, DeviceSortKey, filterDevices, sortDevices } from '@/domain/network';
  import { deviceIcon } from '@/components/ui/icons';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiEmpty from '@/components/ui/UiEmpty.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import RelativeTime from '@/components/ui/RelativeTime.vue';
  import DeviceFilterBar from '@/components/entity/DeviceFilterBar.vue';
  import DeviceChips from '@/components/entity/DeviceChips.vue';
  import DeviceStatusChip from '@/components/entity/DeviceStatusChip.vue';
  import DeviceActions from '@/components/device/DeviceActions.vue';

  const { devices, initialized, isDeviceDisabled } = useHub();
  const filters = ref<DeviceFilters>({ ...DEFAULT_FILTERS });
  const sort = ref<{ key: DeviceSortKey; ascending: boolean }>({ key: 'name', ascending: true });

  const rows = computed(() => sortDevices(filterDevices(devices.value, filters.value, (d) => isDeviceDisabled(d.id)), sort.value.key, sort.value.ascending));

  const columns: { key: DeviceSortKey; label: string; class?: string }[] = [
    { key: 'name', label: 'Device' },
    { key: 'signal', label: 'Signal', class: 'hide-sm' },
    { key: 'battery', label: 'Power', class: 'hide-sm' },
    { key: 'lastSeen', label: 'Last seen', class: 'hide-sm' },
  ];
  function sortBy(key: DeviceSortKey) {
    sort.value = { key, ascending: sort.value.key === key ? !sort.value.ascending : true };
  }
  const ariaSort = (key: DeviceSortKey) => (sort.value.key !== key ? 'none' : sort.value.ascending ? 'ascending' : 'descending');
</script>

<template>
  <div>
    <UiPageHeader title="Device list" :subtitle="`${devices.length} devices`">
      <template #actions>
        <UiButton icon="devices" :to="{ name: RouteName.Devices }">Cards</UiButton>
      </template>
    </UiPageHeader>
    <DeviceFilterBar v-model="filters" :devices="devices" :shown="rows.length" />

    <UiEmpty v-if="initialized && !devices.length" icon="devices" title="No devices yet" text="Pair a Zigbee device with Permit join, or send data from a Wi-Fi device to the hub." />
    <UiEmpty v-else-if="devices.length && !rows.length" icon="search" title="No devices match" text="Try a different search or clear the filters.">
      <UiButton @click="filters = { ...DEFAULT_FILTERS }">Clear filters</UiButton>
    </UiEmpty>
    <div v-else class="nh-table-wrap">
      <table class="nh-table">
        <caption class="sr-only">Devices, sorted by {{ sort.key }}</caption>
        <thead>
          <tr>
            <th v-for="c in columns" :key="c.key" :class="c.class" :aria-sort="ariaSort(c.key)" scope="col">
              <button type="button" @click="sortBy(c.key)">
                {{ c.label }}
                <UiIcon v-if="sort.key === c.key" :name="sort.ascending ? 'chevronUp' : 'chevronDown'" />
              </button>
            </th>
            <th scope="col">Status</th>
            <th scope="col"><span class="sr-only">Actions</span></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="d in rows" :key="d.id" :data-device="d.id">
            <td>
              <RouterLink :to="{ name: RouteName.DevicePage, params: { id: d.id } }" class="nh-dev-link">
                <span class="nh-dev-ic"><UiIcon :path="deviceIcon(d)" /></span>
                <span class="nh-dev-names">
                  <b>{{ d.friendly_name }}</b>
                  <small><code>{{ d.id }}</code></small>
                </span>
              </RouterLink>
            </td>
            <td class="hide-sm nh-chips-cell" colspan="2"><DeviceChips :device="d" /></td>
            <td class="hide-sm"><RelativeTime :value="d.last_seen" /></td>
            <td><DeviceStatusChip :device="d" /></td>
            <td class="nh-row-actions"><DeviceActions :device="d" compact /></td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
  .nh-table-wrap {
    overflow-x: auto;
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
    box-shadow: var(--nh-shadow-1);
    border: var(--nh-border-w) solid var(--nh-border);
  }
  .nh-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
  }
  .nh-table th {
    text-align: left;
    font-weight: 600;
    font-size: 0.78rem;
    color: var(--nh-text-muted);
    padding: 0.6rem 0.75rem;
    border-bottom: var(--nh-border-w) solid var(--nh-border-strong);
    white-space: nowrap;
  }
  .nh-table th button {
    display: inline-flex;
    align-items: center;
    gap: 0.2rem;
    background: none;
    border: 0;
    color: inherit;
    font-weight: inherit;
    cursor: pointer;
    padding: 0.2rem 0;
  }
  .nh-table td {
    padding: 0.5rem 0.75rem;
    border-bottom: var(--nh-border-w) solid var(--nh-border);
    vertical-align: middle;
  }
  .nh-table tbody tr:last-child td {
    border-bottom: 0;
  }
  .nh-table tbody tr:hover {
    background: var(--nh-surface-2);
  }
  .nh-chips-cell {
    white-space: nowrap;
  }
  .nh-chips-cell :deep(.nh-chip) {
    margin-right: 0.3rem;
  }
  .nh-dev-link {
    display: flex;
    align-items: center;
    gap: 0.65rem;
    text-decoration: none;
    min-width: 0;
  }
  .nh-dev-names {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .nh-dev-names small {
    color: var(--nh-text-muted);
  }
  .nh-dev-link:hover b {
    color: var(--nh-accent);
    text-decoration: underline;
  }
  .nh-dev-ic {
    display: grid;
    place-items: center;
    width: 2rem;
    height: 2rem;
    border-radius: var(--nh-radius-m);
    background: var(--nh-accent-soft);
    color: var(--nh-accent);
    flex: none;
  }
  .nh-row-actions {
    text-align: right;
    white-space: nowrap;
  }
  @media (max-width: 760px) {
    .nh-table td,
    .nh-table th {
      padding: 0.5rem;
    }
    .nh-dev-names code {
      font-size: 0.72rem;
    }
  }
</style>
