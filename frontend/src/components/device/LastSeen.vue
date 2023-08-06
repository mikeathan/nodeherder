<script setup>
import { watch, ref, onMounted, onUnmounted } from "vue";
import ElapsedTimer from "../../modules/time-elapsed";


const props = defineProps({
    id: string,
    value: String,
});

let elapsedTimer = new ElapsedTimer();
let lastSeenUpdated = ref(props.value);
let lastSeenElement = ref(null)
watch(
    () => props.value,
    (newlastSeen) => {
        // console.log("Watch props.payload update");
        // elapsedTimer.Format(newlastSeen, (v) => {
        //     lastSeenUpdated.value = v;
        // });
        // lastSeenUpdated.value = elapsedTimer.TimeElapsed;
    },
    { immediate: true }
);

onMounted(() => {
    elapsedTimer.Format2(props.value, lastSeenElement.value);
    console.log("mounted");
});

onUnmounted(() => {
    console.log("unmounted");
    elapsedTimer.dispose();
});
</script>
<template>
    <div title="last update" :ref="'lastSeenElement' + props.id" className="col text-truncate">

    </div>
</template>
