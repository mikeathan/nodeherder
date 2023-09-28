<script setup>
import { useStore } from "vuex";
import { computed, onMounted } from "vue";


const store = useStore();
const automations = computed(() => store.getters.automations);

onMounted(() => {
    store.dispatch('ws/emit', "loadAutomations");
});

</script>

<template>
    <div v-for="(data, name) in automations">

        <p>
        <h3>{{ name }} </h3> : {{ data }}
        </p>
    </div>
</template>
