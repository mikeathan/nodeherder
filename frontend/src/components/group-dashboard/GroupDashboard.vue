<script setup lang="ts">
  import { computed } from 'vue';
  import { store } from '@/store';
  import { DashboardGroup, DashboardGroups, DeviceGroup } from '@/types/settings.type';
  import EntityCard from '../expose/EntityCard.vue';

  const dashboardGroups = computed(() => {
    return store.getters['hub/dashboardGroups']() as DashboardGroups;
  });

  // TESTING just get first expose of thefirst group
  function getDeviceFromGroup(group: DashboardGroup): DeviceGroup {
    const deviceGroup = Object.entries(group.deviceGroup).at(0)?.[1];

    return deviceGroup as DeviceGroup;
  }
  // TESTING
</script>
<style scoped>
  /* .dashboard {
  background-color: var(--p-card-background);
  height: 100vh;
  overflow: auto;
} */
</style>

<template>
  <!-- <div class="grid gap-2 p-2" style="margin: 0; padding: 0.5rem">
    <div class="col-12 sm:col-6 md:col-4 lg:col-3 xl:col-2" v-for="group in dashboardGroups" :key="group.name">
      <div v-for="device in group.deviceGroup" :key="device.deviceId">
        <div v-for="expose in device.exposes" :key="expose">
          <EntityCard :id="device.deviceId" :name="expose" compact />
        </div>
      </div>
    </div>
  </div>  -->

  <div v-for="group in dashboardGroups" :key="group.name" class="mb-4">
    <h4 class="mt-2">{{ group.name }}</h4>
    <div>
      <div v-for="device in group.deviceGroup" :key="device.deviceId" class="grid w-full">
        <div v-for="expose in device.exposes" :key="expose" class="col-12 sm:col-6 md:col-4">
          <EntityCard :id="device.deviceId" :name="expose" compact />
        </div>
      </div>
    </div>

    <div class="flex">
      <div
        class="flex-initial flex align-items-center justify-content-center bg-primary font-bold m-2 px-5 py-3 border-round border-solid">
        Prime
      </div>
      <div
        class="flex-initial flex align-items-center justify-content-center bg-primary font-bold m-2 px-5 py-3 border-round border-solid">
        Prime and PrimeFlex
      </div>
    </div>
    <div class="flex">
      <div
        class="flex-initial flex align-items-center justify-content-center bg-primary font-bold m-2 px-5 py-3 border-round border-solid">
        Prime
      </div>
      <div
        class="flex-initial flex align-items-center justify-content-center bg-primary font-bold m-2 px-5 py-3 border-round border-solid">
        Prime and PrimeFlex
      </div>
      <div
        class="flex-initial flex align-items-center justify-content-center bg-primary font-bold m-2 px-5 py-3 border-round border-solid">
        Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore
        magna aliqua.{' '}
      </div>
    </div>
    --------------------------
    <div class="flex flex-row flex-wrap">
      <div
        class="flex align-items-center justify-content-center w-4rem h-4rem bg-primary font-bold border-round m-2 border-solid">
        1
      </div>
      <div
        class="flex align-items-center justify-content-center w-4rem h-4rem bg-primary font-bold border-round m-2 border-solid">
        2
      </div>
      <div
        class="flex align-items-center justify-content-center w-4rem h-4rem bg-primary font-bold border-round m-2 border-solid">
        3
      </div>
    </div>
    --------------------------
    <div class="flex flex-wrap" style="max-width: 500px">
      <div
        class="flex align-items-center justify-content-center bg-primary font-bold m-2 border-round border-solid"
        style="min-width: 150px; min-height: 100px">
        1
      </div>
      <div
        class="flex align-items-center justify-content-center bg-primary font-bold m-2 border-round border-solid"
        style="min-width: 150px; min-height: 100px">
        2
      </div>
      <div
        class="flex align-items-center justify-content-center bg-primary font-bold m-2 border-round border-solid"
        style="min-width: 150px; min-height: 100px">
        3
      </div>
    </div>
  </div>
</template>
