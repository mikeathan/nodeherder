<script setup>
import { store } from "../../store/index";
import { computed, watch, ref } from "vue";

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
