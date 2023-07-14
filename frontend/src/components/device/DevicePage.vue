<script setup>
import { useStore } from "vuex";
import { computed } from "vue";
import { useRouter } from "vue-router";
import LastSeen from "../device/LastSeen.vue";
import Availability from "../device/Availability.vue";

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
      type: Availability,
      value: device.stats.availability,
    },

    {
      key: "Last seen:",
      type: LastSeen,
      value: device.stats.last_seen,
    },
    {
      key: "Power source:",
      value: device.power_source,
    },
  ];
});

const props = defineProps({
  id: String,
});
</script>
<template>
  <div class="tab-pane fade show active">
    <h1 class="flex-shrink-1">
      {{ id }}
    </h1>

    <dl className="row" v-for="(prop, idx) in displayProps">
      <dt className="col-12 col-md-5">{{ prop.key }}</dt>
      <dd className="col-12 col-md-7" v-if="prop.type == undefined">
        <strong> {{ prop.value }}</strong>
      </dd>
      <dd className="col-12 col-md-7" v-else>
        <component :is="prop.type"></component>
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
