<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";



type Button = { name: string, event: (e: any) => {}, disabled: boolean };
type DropDown = { name: string, items: DropDownItem[], disabled: boolean };
type DropDownItem = { value: string, event: (e: any) => {} };



type PanelItem = Button | DropDown;

export function createButton(name: string, event: (e: any) => {}, disabled: boolean): Button {
    return { name: name, event: event, disabled: disabled };
}



export function createDropDown(name: string, items: DropDownItem[], disabled: boolean): DropDown {
    return { name: name, items: items, disabled: disabled };
}


// function setup() {
//     props.buttons.forEach((item: PanelItem) => {

//         const button = item as Button;
//         if (button !== undefined) {
//             h("button", {
//                 class: "btn btn-light", disabled: button.disabled, onClick(event: any) {
//                     button.event(event)
//                 }
//             }, button.name);


//         } else {
//             const dropDown = item as DropDown;
//             // h("button", {
//             //     class: "btn btn-light", data-bs-toggle: "dropdown", disabled: dropDown.disabled
//             // }, dropDown.name);

//             // h('ul',
//             //     dropDown.items.map((item) => {
//             //         return h('li', { key: item.value,class:"dropdown-item" ,}, item.value)
//             //     })
//             // )
//         }
//     });
// }



const props = defineProps({
    buttons: {
        type: Object as PropType<Array<Button>>,
        default: [],
        required: true
    },
});


</script>

<template>
    <div class="row pb-3">
        <form class="container">

            <div v-for="item in props.buttons">
                <button class="btn btn-light " type="button" @click="item.event" :disabled="item.disabled">{{ item.name
                    }}</button>
            </div>

            <div v-for="item in props.dropDowns">
                <button class="btn btn-light " type="button" data-bs-toggle="dropdown" :disabled="item.disabled">{{
                item.name }}</button>


                <ul class="dropdown-menu">
                    <li v-for="value in item.values">
                        <a @click="addStep(operator as NumericOperator)" class="dropdown-item" data-toggle="dropdown">
                            {{ value }}</a>
                    </li>
                </ul>
            </div>

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