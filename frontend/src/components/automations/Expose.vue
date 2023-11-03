<script setup>
import { useStore } from "vuex";
import { computed, ref, onBeforeMount } from "vue";
import Condition from "./Condition"
import { Expose } from "../../models/automation"
import { ExposeCondition } from "../../models/automation"

const exposes = ref({})


const props = defineProps({
    expose: Object,
});


function addCondition() {

    var newcondition = new ExposeCondition();
    exposes.value[selectedExpose.value].addCondition(newcondition)

    console.log("add condition")
}

</script>
<template>
    <div class="container-fluid p-0 h-100">
        <h4>Condition</h4>

        <div class="col">
            <div v-for=" condition in props.expose.Conditions">
                <Condition :Expose="props.expose" :Condition="condition"></Condition>
            </div>
        </div>
        <div class="col-3 pt-3">
            <input type="button" class="btn btn-secondary text-nowrap" value="Add" @click="addCondition()" />
        </div>
    </div>
</template>
