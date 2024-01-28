<script setup>
import { watch, ref, onMounted, onUnmounted } from "vue";
import ElapsedTimer from "../../modules/time-elapsed";

const props = defineProps({
    timestamp: String,
});

let elapsedTimer = null;
let lastSeenElement = ref(null)
watch(
    () => props.timestamp,
    (newlastSeen) => {
        if (lastSeenElement.value == undefined) {
            return;
        }
        //console.log("Watch props.payload update");
        elapsedTimer.Format(newlastSeen);
    },
    { immediate: true }
);

onMounted(() => {
    elapsedTimer = new ElapsedTimer(lastSeenElement.value);
    elapsedTimer.Format(props.timestamp);
});

onUnmounted(() => {
    elapsedTimer.dispose();
});
</script>
<template>
    <div :title="'last update ' + timestamp" :ref="el => { lastSeenElement = el }" className="col text-truncate">

    </div>
</template>
