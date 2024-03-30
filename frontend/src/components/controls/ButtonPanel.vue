<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive, h, VNode } from "vue";
import Dropdown from "@/components/controls/Dropdown.vue";
import Button from "@/components/controls/Button.vue";

import { ButtonPanelType, ButtonType, isDropdown, DropDownType } from "@/types/controls.type";
import { PanelComponents } from "@/mixins/usePanelComponents";
import { KeyyValuePair } from "@/types/types";

function HelloWorld({ }) {
    return h(
        'div',
        {
            style: {
                color: 'red',

            },
        },
        'Hello World'
    )
}
function Panel() {

    // return h("button", {
    //     class: "btn btn-light",
    //     disabled: false
    // }
    //     , "TEST BUTTON");

    //     const vnode = h('div', { id: 'foo' }, [])

    // vnode.type // 'div'
    // vnode.props // { id: 'foo' }
    // vnode.children // []
    // vnode.key // null

    let vNodes: VNode[] = [];
    props.buttons.forEach((item: ButtonPanelType) => {

        if (isDropdown(item)) {
            const dropDown = item as DropDownType;
            const node = h("Dropdown", {
                className: 'btn-light',
                disabled: dropDown.disabled,
                items: dropDown.items
            }, dropDown.name);

            vNodes.push(node);

        } else {
            const button = item as ButtonType;
            const node = h("button", {
                class: "btn btn-light",
                disabled: button.disabled,
                onClick(event: any) {
                    button.click(event)
                }
            }, button.name);

            vNodes.push(node);
        }

    });
    return vNodes;

};

const props = defineProps({
    buttons: {
        type: Object as PropType<Array<ButtonPanelType>>,
        default: [],
        required: true
    },
});


</script>

<template>
    <HelloWorld></HelloWorld>
    <Panel></Panel>
    <div class="row pb-3">
        <!-- <form class="container">

            <form class="container" v-for="item in buttonComponents">
                {{ item }}

            </form>
        </form> -->
    </div>
</template>