<script setup>
import { useStore } from "vuex";
import { computed, watch, ref } from "vue";

const store = useStore();
const connected = computed(() => {
    return store.getters["wsclient/isconnected"];
});
const color = ref("red")
// let connectionState = { true, "green", false, "red"}
watch(
    () => connected.value,
    () => {
        console.log("status change", connected.value)
        if (connected.value) {
            color.value = "green"
        } else {
            color.value = "red"
        }
    },
    { immediate: true }
);




</script>

<template>
    <span :style="'color:' + color + ';'">
        <i class="fa fa-circle fa-xs"></i>
    </span>
</template>
