<script setup>
import { useStore } from "vuex";
import { onBeforeMount, onMounted, ref } from "vue";
import { Operators, ExposeCondition, Expose } from "../../models/automation"

const props = defineProps({
    Expose: Object,
    Condition: Object
});

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
                <select id="selectOperators" style="text-align:center;" class="form-control"
                    v-model="props.Condition.Operator">
                    <option v-for="operator in Operators" :value="operator.value" :key="operator.value">
                        {{ operator.text }}
                    </option>
                </select>
            </div>
            <div class="col">
                <input type="text" style="text-align:center;" class="form-control" placeholder="Condition value"
                    onfocus="this.placeholder = ''" onblur="this.placeholder='Condition value'"
                    v-model="props.Condition.Value" />
            </div>

        </div>
    </div>
</template>
