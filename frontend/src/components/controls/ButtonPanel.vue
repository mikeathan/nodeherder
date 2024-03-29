<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive, h } from "vue";
import Dropdown from "@/components/controls/Dropdown.vue";
import Button from "@/components/controls/Button.vue";

import { ButtonPanelType, isDropdown, DropDownType } from "@/types/controls.type";
import { PanelComponents } from "@/mixins/usePanelComponents";



function setup() {
    // buttonComponents.value = []
    ; props.buttons.forEach((item: ButtonPanelType) => {

        if (isDropdown(item)) {
            const dropDown = item as DropDownType;
            // h("button", {
            //     class: "btn btn-light", disabled: button.disabled, onClick(event: any) {
            //         button.event(event)
            //     }
            // }, button.name);


        } else {
            // h("button", {
            //     class: "btn btn-light", data-bs-toggle: "dropdown", disabled: dropDown.disabled
            // }, dropDown.name);

            // h('ul',
            //     dropDown.items.map((item) => {
            //         return h('li', { key: item.value,class:"dropdown-item" ,}, item.value)
            //     })
            // )
        }
    });
}


const buttonComponents = ref<string[]>([]);
const props = defineProps({
    buttons: {
        type: Object as PropType<Array<ButtonPanelType>>,
        default: [],
        required: true
    },
});


</script>

<template>
    <div class="row pb-3">
        <form class="container">

            <form class="container" v-for="item in props.buttons">
                <button class="btn btn-light " type="button" @click="saveAction"
                    :disabled="!isSaveEnabled">Save</button>
                <button class="btn btn-light btn" type="button" @click="removeAction">Delete</button>

                <Dropdown :items="dropDownitems" class-name="btn-light" :disabled="action.id == ''">Add Operation
                </Dropdown>

            </form>
        </form>
    </div>
</template>