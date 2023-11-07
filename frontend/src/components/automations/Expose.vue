<script setup>
import Conditions from "./Conditions"
import { Condition } from "../../models/automation"
import { onMounted, onUpdated } from "vue";

const props = defineProps({
    expose: Object,
});

// not used
function add(event) {
    console.log("Expose add condition ", event);
    props.expose.addCondition(event)
}

// not used
function remove(event) {
    console.log("Expose remove condition ", event);
    props.expose.remove(event)
}
onUpdated(() => {
    console.log("EXPOSE onUpdated ");
});
</script>
<template>
    <div class="container-fluid p-0 h-100">

        <h4>Condition</h4>
        {{ props.expose.Name }}
        <div class="col">
            <Conditions :expose="props.expose" :exposeName="props.expose.Name">
            </Conditions>
        </div>
    </div>
</template>
