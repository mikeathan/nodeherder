<script setup lang="ts">
import { computed, ref } from "vue";
import { store } from "../../store/index";
import { useRouter } from "vue-router";
import DeviceAbout from "./DeviceAbout.vue";
import DeviceExposes from "./DeviceExposes.vue";
import DeviceSettings from "./DeviceSettings.vue";
import DeviceMetrics from "./DeviceMetrics.vue";
import { deviceTabComponents } from '../../mixins/useTabComponents'
import { Device } from "@/types/device";

const props = defineProps({
  id: {
    type: String,
    required: true
  }
});

const deviceExist = computed<boolean>(() => {
  return store.getters["devices/exists"](props.id);
});

const device = computed<Device>(() => {
  return store.getters["devices/find"](props.id);
});

const previousPage = computed(() => {
  const back = useRouter().options.history.state.back;
  if (back != undefined) {
    return back
  }

  return useRouter().push("/");
});


</script>
<template>
  <div v-if="deviceExist">

    <div className="d-flex flex-row">
      <div class="align-self-center me-3">
        <RouterLink :to="`${previousPage}`">
          <i class="fa fa-arrow-left fa-xl" aria-hidden="true"></i>
        </RouterLink>
      </div>
      <div class="h3 align-self-center">
        {{ device.friendly_name }}
      </div>
    </div>
    <div class="col-12 col-md-9 ">
      <Tabs value="0">
        <TabList>
          <Tab v-for="tab in deviceTabComponents" :key="tab.title" :value="tab.value">{{ tab.title }}</Tab>
        </TabList>
        <TabPanels>
          <TabPanel v-for="tab in deviceTabComponents" :key="tab.value" :value="tab.value">
            <component :is="tab.content" v-bind="{ id: props.id }"></component>
          </TabPanel>
        </TabPanels>
      </Tabs>
    </div>
  </div>
</template>
