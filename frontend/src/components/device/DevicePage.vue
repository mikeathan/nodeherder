<script setup lang="ts">
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import Tabs from '../controls/Tabs.vue'
import Tab from '../controls/Tab.vue'
import DeviceInfo from "./DeviceInfo.vue";
import DeviceExposes from "./DeviceExposes.vue";

const props = defineProps({
  id: { type: String, required: true }
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
  <div class="col-12 col-md-9 panel">
    <Tabs>
      <Tab active="true" title="About">
        <DeviceInfo :id="props.id"></DeviceInfo>
      </Tab>
      <Tab title="Exposes">
        <DeviceExposes :id="props.id"></DeviceExposes>
      </Tab>
    </Tabs>
  </div>
</template>
