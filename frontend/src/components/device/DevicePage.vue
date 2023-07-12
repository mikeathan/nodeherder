<script setup>
import { useStore } from "vuex";
import { computed } from "vue";

const store = useStore();
const device = computed(() => store.getters.findDevice(props.id));
const props = defineProps({
  id: String,
});


// const User = {
//   // make sure to add a prop named exactly like the route param
//   props: ['id'],
//   template: '<div>User {{ id }}</div>'
// }
// const routes = [{ path: '/user/:id', component: User, props: true }]
const displayProps = [
  {
    // translationKey: "Friendly name",
    // render: () => (
    //   <dd className="col-12 col-md-7">
    //     <strong>{device.id}</strong>
    //   </dd>
    // ),
    // translationKey: "Connection Type",
    // render: () => (
    //   <dd className="col-12 col-md-7">
    //     <strong>{device.conn}</strong>
    //   </dd>
    // ),
    // translationKey: "Last seen",
    // render: () => (
    //   <dd className="col-12 col-md-7">
    //     <strong> WIP</strong>
    //   </dd>
    // ),
    // translationKey: "Availability",
    // render: () => (
    //   <dd className="col-12 col-md-7">
    //     <strong> WIP</strong>
    //   </dd>
    // ),
    // translationKey: "Availability",
    // render: () => (
    //   <dd className="col-12 col-md-7">
    //     <strong> WIP</strong>
    //   </dd>
    // ),
    // translationKey: "Power source",
    // render: () => (
    //   <dd className="col-12 col-md-7">
    //     <strong> WIP</strong>
    //   </dd>
    // ),
  },
];
</script>
<template>
  <h1>
    {{ id }}
  </h1>
  <div>availability: {{ device.stats.availability }}</div>

  <!-- <dl className="row" v-for="(key, value) in displayProps">
    <dt className="col-12 col-md-5">{key}</dt>
  </dl> -->
  <!-- add device control -->
</template>
