<script setup lang="ts">
import { computed, ref } from "vue";
import { store } from "../../store/index";
import { useRouter } from "vue-router";
import Tabs from '../controls/Tabs.vue'
import Tab from '../controls/Tab.vue'
import DeviceAbout from "./DeviceAbout.vue";
import DeviceExposes from "./DeviceExposes.vue";
import { Device } from "@/types/device";

const props = defineProps({
  id: { type: String, required: true }
});

const deviceExist = computed<boolean>(() => {
  return store.getters["devices/exists"](props.id);
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
  <div v-if="deviceExist" class="col-12 col-md-9 panel">
    <Tabs>
      <Tab active="true" title="About">
        <DeviceAbout :id="props.id"></DeviceAbout>
      </Tab>
      <Tab title="Exposes">
        <DeviceExposes :id="props.id"></DeviceExposes>
      </Tab>
    </Tabs>
  </div>
</template>
