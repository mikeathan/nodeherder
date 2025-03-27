<script setup lang="ts">
  import { computed } from 'vue';
  import { Devices } from '@/types/device';
  import { store } from '../../store/index';
  import LastSeen from '../device/LastSeen.vue';
  import { getPowerSourceValue } from '@/contracts/device';
  import Icon from '../controls/Icon.vue';
  import { getOfflineIcon } from '@/modules/formatters/device.formatter';

  const devices = computed(() => store.getters['hub/listAllDevices']() as Devices);
</script>
<style scoped>
  .last-seen {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
</style>

<template>
  <div class="device-list-container">
    <div class="header">
      <h2>Devices</h2>
    </div>
    <DataTable
      :value="devices"
      paginator
      :rowsPerPageOptions="[10, 20, 50]"
      :first="0"
      :rows="20"
      sortField="friendly_name"
      :sortOrder="1">
      <Column field="friendly_name" header="Name" sortable>
        <template #body="slotProps">
          <RouterLink :to="`/devicepage/${slotProps.data.id}`" class="link">
            {{ slotProps.data.friendly_name }}
          </RouterLink>
        </template>
      </Column>
      <Column field="id" header="IEEE Address" />
      <Column header="Last seen" style="width: 200px; min-width: 200px">
        <template #body="slotProps">
          <span class="last-seen">
            <div v-if="slotProps.data.availability === 'online'">
              <LastSeen :timestamp="slotProps.data.last_seen" />
            </div>
            <div v-else>
              <Icon :icon="getOfflineIcon()" />
            </div>
          </span>
        </template>
      </Column>
      <Column header="Power">
        <template #body="slotProps">
          <PowerSource :power_source="slotProps.data.power_source" :value="getPowerSourceValue(slotProps.data)" />
        </template>
      </Column>
      <Column header="Actions">
        <template #body="slotProps">
          <DeviceControl :id="slotProps.data.id" />
        </template>
      </Column>
    </DataTable>
  </div>
</template>
