<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";

type ButtonItem = { name: string, event: any, disabled: boolean };
type DropDownItem = { names: string[], event: any, disabled: boolean };
type PanelItem = ButtonItem | DropDownItem;

function createButton(): ButtonItem {
    return null;
}

function createDropDown(): DropDownItem {
    return null;
}
const props = defineProps({
    items: {
        type: Object as PropType<Array<PanelItem>>,
        default: [],
        required: true
    },
});


</script>

<template>
    <div class="row pb-3">
        <form class="container">
            <button class="btn btn-light " type="button" @click="saveAction" :disabled="!isSaveEnabled">Save</button>
            <button class="btn btn-light btn" type="button" @click="removeAction">Delete</button>
            <button type="button" class="btn btn-light" data-bs-toggle="dropdown" :disabled="action.id == ''">Add
                Operation</button>
            <ul class="dropdown-menu">
                <li v-for="operator in NumericOperators">
                    <a @click="addStep(operator as NumericOperator)" class="dropdown-item" data-toggle="dropdown">
                        {{ operator }}</a>
                </li>
            </ul>
        </form>
    </div>
</template>