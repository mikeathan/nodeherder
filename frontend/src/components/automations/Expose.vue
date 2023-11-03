<script setup>
import Condition from "./Condition"
import { ExposeCondition } from "../../models/automation"

const props = defineProps({
    expose: Object,
});

function addCondition() {
    var newcondition = new ExposeCondition();
    props.expose.addCondition(newcondition)
}

function removeLastCondition() {
    props.expose.removeLastCondition()
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
            <div class="btn-group" role="group">
                <input type="button" class="btn btn-secondary text-nowrap" value="Add" @click="addCondition()" />
                <input type="button" class="btn btn-secondary text-nowrap" value="Remove" @click="removeLastCondition()"
                    :disabled="props.expose.hasConditions() == false" />
            </div>
        </div>
    </div>
</template>
