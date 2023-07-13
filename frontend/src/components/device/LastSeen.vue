<script setup>
import DeviceFormatter from "../../modules/sensors/device-formatter";

const props = defineProps({
    lastSeen: String,
});
let formatter = new DeviceFormatter();
let lastSeenUpdated = ref("");

watch(
    () => props.lastSeen,
    (newpayload) => {
        console.log("Watch props.payload update");
        formatter.formatPayload(newpayload.stats, (v) => {
            lastSeenUpdated.value = v;
        });
        lastSeenUpdated.value = formatter.lastSeen;
    },
    { immediate: true }
);

onMounted(() => {
    console.log("mounted");
});

onUnmounted(() => {
    console.log("unmounted");
    formatter.dispose();
});
</script>
<template>
    <div title="last update" className="col text-truncate">
        {{ lastSeenUpdated }}
    </div>
</template>
