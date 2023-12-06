<script setup>
import { useStore } from "vuex";
import { computed, watch, ref } from "vue";

const store = useStore();
const connected = computed(() => {
    return store.getters["ws/isconnected"];
});
const color = ref("red")
const connectionState = { true: "green", false: "red", null: 'white' }
watch(
    () => connected.value,
    () => {
        color.value = connectionState[connected.value]
    },
    { immediate: true }
);




</script>

<template>
    <span :style="'color:' + color + ';'">
        <i class="fa fa-circle fa-xs"></i>
    </span>
</template>
