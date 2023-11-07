<script setup>
import Conditions from "./Conditions"
import { Condition } from "../../models/automation"
import { onMounted } from "vue";

const props = defineProps({
    expose: Object,
});

function add(event) {
    console.log("Expose add condition ", event);
    props.expose.addCondition(event)
    // for (var i = 0; i < props.expose.Conditions.length; i++) {
    //     console.log(props.expose.Conditions[i]);
    // }
}

function remove(event) {
    console.log("Expose remove condition ", event);
    props.expose.remove(event)

    // for (var i = 0; i < props.expose.Conditions.length; i++) {
    //     console.log(props.expose.Conditions[i]);
    // }
}

</script>
<template>
    <div class="container-fluid p-0 h-100">

        <h4>Condition</h4>
        <div class="col">
            <Conditions :expose="props.expose" :exposeName="props.expose.Name" @add="add($event)" @remove="remove($event)">
            </Conditions>
        </div>
    </div>
</template>
