<script setup>
import MqqtImage from "../../../../assets/mqttorg-graphics-master/png/mqtt-hor-neg.png"
import HttpImage from "../../../../assets/HTTP_logo.svg.png"
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
