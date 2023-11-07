<script setup>
import Condition from "./Condition"
import { ExposeCondition } from "../../models/automation"
import { onMounted } from "vue";

const props = defineProps({
    expose: Object,
});

function addCondition(event) {

    props.expose.addCondition(event.value)
}

function remove(event) {

    props.expose.remove(event.value)
}

</script>
<template>
    <div class="container-fluid p-0 h-100">

        <h4>Condition</h4>
        <div class="col">
            <Condition :exposeName="props.expose.Name">
            </Condition>
        </div>
        <!-- <div class="col">
            <Condition :name="props.expose.Name" @add="addCondition($event)" :condition="new ExposeCondition()"
                @remove="remove($event)">
            </Condition>
        </div>
        <div class="col">
            <div v-for=" condition in props.expose.Conditions">

                <Condition :name="props.expose.Name" :condition="condition" @add="addCondition($event)"
                    @remove="remove($event)"></Condition>
            </div>
        </div> -->

        <!-- <div class="col-3 pt-3">
            <div class="btn-group" role="group">
                <input type="button" class="btn btn-secondary text-nowrap" value="Add" @click="addCondition()" />
                <input type="button" class="btn btn-secondary text-nowrap" value="Remove" @click="removeLastCondition()"
                    :disabled="props.expose.hasConditions() == false" />
            </div>
        </div> -->
    </div>
</template>
