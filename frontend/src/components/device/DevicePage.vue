<script setup>
import { useStore } from "vuex";
import { computed } from "vue";
import { useRouter } from "vue-router";
import LastSeen from "../device/LastSeen.vue";
import PowerSource from "../device/PowerSource.vue";

const previousPage = computed(() => {
  return useRouter().options.history.state.back;
});

const store = useStore();
const displayProps = computed(() => {
  const device = store.getters.findDevice(props.id);
  if (device == undefined) {
    return [];
  }
  return [
    {
      key: "Friendly name:",
      value: device.id,
    },
    {
      key: "Connection Type:",
      value: device.conn,
    },
    {
      key: "Availability:",
      value: device.stats.availability,
    },

    {
      key: "Last seen:",
      type: LastSeen,
      props: {
        value: device.stats.last_seen,
      },
    },
    {
      key: "Power source:",
      type: PowerSource,
      props: {
        power_source: device.power_source,
        value: device.stats.battery,
      },
    },
  ];
});

const props = defineProps({
  id: String,
});
</script>
<template>
  <div class="tab-pane fade show active">
    <div className="d-flex flex-row">
      <div class="align-self-center">
        <i class="fa fa-times fa-lg" aria-hidden="true"></i>
      </div>
      <div class="h1 align-self-center">
        {{ id }}
      </div>
    </div>
    <dl className="row" v-for="(prop, idx) in displayProps">
      <dt className="col-12 col-md-5">{{ prop.key }}</dt>
      <dd className="col-12 col-md-7" v-if="prop.type == undefined">
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
      <button class="btn btn-danger" title="Remove device">
        <i class="fa fa-trash"></i>
      </button>
    </div>
  </div>
  <!-- todo -->
  <RouterLink :to="`${previousPage}`">
    <i class="fa fa-arrow-left" aria-hidden="true"></i>
  </RouterLink>
</template>
