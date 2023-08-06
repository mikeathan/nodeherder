<script setup>
import { watch, ref, onMounted, onUnmounted } from "vue";
import ElapsedTimer from "../../modules/time-elapsed";

const props = defineProps({
    value: String,
});

let elapsedTimer = new ElapsedTimer();
let lastSeenUpdated = ref(props.value);
let lastSeenElement = ref(null)
watch(
    () => props.value,
    (newlastSeen) => {
        console.log("Watch props.payload update");
        elapsedTimer.SetTimestamp(newlastSeen, (v) => {
            lastSeenUpdated.value = v;
        });
        lastSeenUpdated.value = elapsedTimer.TimeElapsed;
    },
    { immediate: true }
);

onMounted(() => {
    console.log("mounted");


});

onUnmounted(() => {
    console.log("unmounted");
    elapsedTimer.dispose();
});
</script>
<template>
    <div title="last update" ref="lastSeenElement" className="col text-truncate">
        {{ lastSeenUpdated }}
    </div>
</template>
