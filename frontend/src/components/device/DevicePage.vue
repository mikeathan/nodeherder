<script setup>
import { useStore } from "vuex";
import { computed } from "vue";
import { useRouter } from "vue-router";
import LastSeen from "../device/LastSeen.vue";
import PowerSource from "../device/PowerSource.vue";
import ConnectionType from "../device/ConnectionType.vue";
const previousPage = computed(() => {
  var back = useRouter().options.history.state.back;
  if (back == undefined) {
    back = useRouter().push("/");
  }
  return back;
});

const store = useStore();
const friendlyName = computed(() => {

  const d = store.getters["devices/find"](props.id);
  if (d == undefined) {
    return "";
  }
  return d.friendlyName;
});

const displayProps = computed(() => {
  const device = store.getters["devices/find"](props.id);
  if (device == undefined) {
    return [];
  }
  return [

    {
      key: "Id:",
      value: device.id,
    },
    {
      key: "Friendly name:",
      value: device.friendlyName,
    },
    {
      key: "Description:",
      value: device.description,
    },
    {
      key: "Availability:",
      value: device.properties.availability,
    },
    {
      key: "Last seen:",
      type: LastSeen,
      props: {
        timestamp: device.properties.last_seen,
      },
    },
    {
      key: "Power source:",
      type: PowerSource,
      props: {
        power_source: device.power_source,
        value: device.properties.battery,
      },
    },
    {
      key: "Connection Type:",
      type: ConnectionType,
      props: {
        type: device.connection_type,
      },
    },
  ];
});

const props = defineProps({
  id: String,
});
function showDialog(message) {
  alert(message);
}
</script>
<template>
  <div class="panel">
    <div className="panel-header d-flex flex-row">
      <div class="align-self-center me-3">
        <RouterLink :to="`${previousPage}`">
          <i class="fa fa-arrow-left fa-xl" aria-hidden="true"></i>
        </RouterLink>
      </div>
      <div class="h1 align-self-center">
        {{ friendlyName }}
      </div>
    </div>

    <div>
      <dl className="row align-self-center" v-for="(prop, idx) in displayProps">
        <dt className="col-12 col-md-5">{{ prop.key }}</dt>
        <dd className="col-12 col-md-7 " v-if="prop.type == undefined">
          <div title="last update" className="col text-truncate">
            {{ prop.value }}
          </div>
        </dd>
        <dd className="col-12 col-md-7" v-else>
          <component :is="prop.type" v-bind="prop.props"></component>
        </dd>
      </dl>

      <!-- todo -->
      <div class="btn-group btn-group-sm" role="group">
        <button class="btn btn-danger" title="Remove device" @click="showDialog('Not implemented')">
          <i class="fa fa-trash"></i>
        </button>
      </div>
    </div>
  </div>
</template>
