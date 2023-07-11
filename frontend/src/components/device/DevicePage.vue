<script setup>
const props = defineProps({
  device: Object,
});
const displayProps = [
  {
    translationKey: "Friendly name",
    render: () => (
      <dd className="col-12 col-md-7">
        <strong>{device.id}</strong>
      </dd>
    ),
    translationKey: "Connection Type",
    render: () => (
      <dd className="col-12 col-md-7">
        <strong>{device.conn}</strong>
      </dd>
    ),
    translationKey: "Last seen",
    render: () => (
      <dd className="col-12 col-md-7">
        <strong> WIP</strong>
      </dd>
    ),
    translationKey: "Availability",
    render: () => (
      <dd className="col-12 col-md-7">
        <strong> WIP</strong>
      </dd>
    ),
    translationKey: "Availability",
    render: () => (
      <dd className="col-12 col-md-7">
        <strong> WIP</strong>
      </dd>
    ),
    translationKey: "Power source",
    render: () => (
      <dd className="col-12 col-md-7">
        <strong> WIP</strong>
      </dd>
    ),
  },
];
</script>
<template>
  <dl className="row" v-for="(key, value) in displayProps">
    <dt className="col-12 col-md-5">{key}</dt>
  </dl>
  <!-- add device control -->
</template>
