<script setup lang="ts">
import { store } from "../../store/index";
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import LastSeen from "../device/LastSeen.vue";
import PowerSource from "../device/PowerSource.vue";
import ConnectionType from "../device/ConnectionType.vue";
import RenameDeviceDialog from "../dialogs/RenameDeviceDialog.vue";
import { Device } from "@/types/device";
import Tabs from '../controls/Tabs.vue'
import Tab from '../controls/Tab.vue'
import DeviceInfo from "../device/DeviceInfo.vue";
const props = defineProps({
  id: String,
});

const previousPage = computed(() => {
  const back = useRouter().options.history.state.back;
  if (back != undefined) {
    return back
  }

  return useRouter().push("/");
});


const devicePageTabs = computed(() => {
  const device = store.getters["devices/find"](props.id) as Device;
  if (device == undefined) {
    return [{
      title: "Info",
      enabled: true,
      text: "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce gravida purus vitae vulputate commodo."
    }, {
      title: "Something else",
      enabled: false,
      text: "Cras scelerisque, dolor vitae suscipit efficitur, risus orci sagittis velit, ac molestie nulla tortor id augue."
    }
    ];
  }
  return [
  ]
});

const showDialog = ref(false)
const device = computed(() => {
  return store.getters["devices/find"](props.id);
});


</script>
<template>
  <!-- <div>
    <Tabs>
      <Tab active="true" title="First Tab">
        Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce gravida purus vitae vulputate commodo.
      </Tab>
      <Tab title="Second Tab">
        Cras scelerisque, dolor vitae suscipit efficitur, risus orci sagittis velit, ac molestie nulla tortor id augue.
      </Tab>
      <Tab title="Third Tab">
        Morbi posuere, mauris eu vehicula tempor, nibh orci consectetur tortor, id eleifend dolor sapien ut augue.
      </Tab>
      <Tab title="Fourth Tab">
        Aenean varius dui eget ante finibus, sit amet finibus nisi facilisis. Nunc pellentesque, risus et pretium
        hendrerit.
      </Tab>
    </Tabs>
  </div> -->

  <div>
    <Tabs>
      <Tab active="true" title="info">
        <component :is="DeviceInfo" v-model="props"></component>
      </Tab>
      <Tab title="Something else">
        <component :is="DeviceInfo" v-model="props"></component>
      </Tab>
    </Tabs>
  </div>
</template>
