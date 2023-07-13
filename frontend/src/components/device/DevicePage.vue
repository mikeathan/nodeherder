<script setup>
import { useStore } from "vuex";
import { computed } from "vue";
import { useRouter } from "vue-router";

const previousPage = computed(() => {
  return useRouter().options.history.state.back;;
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
  <div>

    <span>
      <RouterLink :to="`${previousPage}`">
        <i class="fa fa-arrow-left" aria-hidden="true"></i>
      </RouterLink>
    </span>

    <span>
      <h1> {{ id }}</h1>
    </span>
  </div>


  <dl className="row" v-for="(prop, idx) in displayProps">
    <dt className="col-12 col-md-5">{{ prop.key }}</dt>
    <dd className="col-12 col-md-7">
      <strong> {{ prop.value }}</strong>
    </dd>
  </dl>
</template>
