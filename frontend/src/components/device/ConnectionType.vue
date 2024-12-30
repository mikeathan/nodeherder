<script setup>
  import MqqtImage from '../../assets/images/transport/mqtt-logo.png';
  import HttpImage from '../../assets/images/transport/http-logo.png';
  const props = defineProps({
    type: String,
  });
</script>
<template>
  <div title="ConnectionType" v-if="type === 'mqtt'">
    <img :src="MqqtImage" width="100" />
  </div>
  <div title="ConnectionType" v-else-if="type === 'http'">
    <img :src="HttpImage" width="80" />
  </div>
  <div title="ConnectionType" v-else>
    {{ type }}
  </div>
</template>
