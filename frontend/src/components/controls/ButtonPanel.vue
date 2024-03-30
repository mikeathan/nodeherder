<script setup lang="ts">
import { PropType, h, VNode } from "vue";
import { ButtonPanelType, ButtonType, isDropdown, DropDownType } from "@/types/controls.type";

function Panel() {
    let vNodes: VNode[] = [];
    props.buttons.forEach((item: ButtonPanelType) => {
        const node = isDropdown(item) ? createDropdown(item as DropDownType) : createButton(item as ButtonType);
        vNodes.push(node);
    });
    return vNodes;

};

function createDropdown(dropDown: DropDownType): VNode {
    return h("Dropdown", {
        className: 'btn-light',
        disabled: dropDown.disabled,
        items: dropDown.items
    }, dropDown.name);
}

function createButton(button: ButtonType): VNode {
    return h("button", {
        class: "btn btn-light",
        disabled: button.disabled,
        onClick(event: any) {
            button.click(event)
        }
    }, button.name);
}

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
            <Panel></Panel>
        </form>
    </div>
</template>