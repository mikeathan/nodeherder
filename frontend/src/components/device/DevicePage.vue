<script setup>
import { useStore } from "vuex";
import { computed } from "vue";

const store = useStore();
const device = computed(() => store.getters.findDevice(props.id));
const props = defineProps({
  id: String,
});


const User = {
  // make sure to add a prop named exactly like the route param
  props: ['id'],
  template: '<div>User {{ id }}</div>'
}
// const routes = [{ path: '/user/:id', component: User, props: true }]
const displayProps = [
  {
    key: "Friendly name",
    value: device.id
  }
  ,
  {
    key: "Connection Type",
    value: device.conn
  },
  {
    key: "Last seen",
    value: device.stats.lastSeen
  },
  {
    key: "Availability",
    value: device.stats.availability
  }, {
    key: "Power source",
    value: device.power_source
  },
  {
    key: "Power source",
    value: device.power_source
  }

  // translationKey: "Last seen",
  // render: () => (
  //   <dd className="col-12 col-md-7">
  //     <strong> WIP</strong>
  //   </dd>

];
</script>
<template>
  <h1>
    {{ id }}
  </h1>
  <div>Availability: {{ device.stats.availability }}</div>

  <dl className="row" v-for="(key, value) in displayProps">
    <dt className="col-12 col-md-5">{{ key }}</dt>
    <dd className="col-12 col-md-7">
      <strong> {{ value }}</strong>
    </dd>
  </dl>
</template>
