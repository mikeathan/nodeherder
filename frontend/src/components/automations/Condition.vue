<script setup>
import { useStore } from "vuex";
import { computed } from "vue";
import { Operators } from "../../models/automation"


const props = defineProps({
    Expose: Object
});

function addCondition() {
    console.log("add condition");
    //props.Expose.addCondition()
}
</script>

<template>
    <div class="container-fluid p-0 h-100">
        <h4>Condition</h4>
        <div class="row w-50">
            <div class="col">
                <input type="text" class="form-control" placeholder="Condition name" onfocus="this.placeholder = ''"
                    onblur="this.placeholder='Condition name'" v-model="props.Expose.Name" disabled />
            </div>
            <div class="col">
                <select id="selectOperators" style="text-align:center;" class="form-control">
                    <option v-for="operator in Operators" :value="operator.value" :key="operator.value">
                        {{ operator.text }}
                    </option>
                </select>
            </div>
            <div class="col">
                <input type="text" style="text-align:center;" class="form-control" placeholder="Condition value"
                    onfocus="this.placeholder = ''" onblur="this.placeholder='Condition value'" v-model="props.Expose.Value"
                    p />
            </div>
            <div class="col">
                <input type="button" class="btn btn-secondary text-nowrap" value="Add" @click="addCondition()" />
            </div>

        </div>
    </div>
</template>
