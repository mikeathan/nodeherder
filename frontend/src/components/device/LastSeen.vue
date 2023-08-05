<script setup>
import { watch, ref, onMounted, onUnmounted } from "vue";
import ElapsedTimer from "../../modules/time-elapsed";

const props = defineProps({
    value: String,
});

let elapsedTimer = new ElapsedTimer();
let lastSeenUpdated = ref(props.value);

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
    <div title="last update" className="col text-truncate">
        {{ lastSeenUpdated }}
    </div>
</template>
const nodes = document.querySelectorAll('.timeago');

// use render method to render nodes in real time
render(nodes, 'zh_CN');

// render with opts
// render(nodes, 'en_US', { minInterval: 3 });

// cancel all real-time render task
cancel();