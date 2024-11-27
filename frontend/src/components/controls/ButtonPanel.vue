<script setup lang="ts">
import { PropType, h, VNode } from "vue";
import { ButtonPanelType, ButtonType, isDropdown, DropDownType } from "@/types/controls.type";


const props = defineProps({
    buttons: {
        type: Object as PropType<Array<ButtonPanelType>>,
        default: [],
        required: true
    },
});

function Panel() {
    return props.buttons.map((item: ButtonPanelType) => isDropdown(item)
        ? createDropdown(item as DropDownType)
        : createButton(item as ButtonType));
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
        onClick: (event: any) => {
            event.preventDefault();
            try {
                button.click(event);
            } catch (error) {
                console.error('Error during button click:', error);
            }
        }
    }, button.name);
}
function createButton2(button: ButtonType): VNode {
    functional 
}
</script>

<template>
    <div class="row pb-3">
        <form class="container">
            <Button label="Primary" variant="text" raised />

            <Button
            v-for="buttonInfo in props.buttons"
            :key="buttonInfo.name"
            :label="buttonInfo.name"
            :disabled="buttonInfo.disabled"
            @click="buttonInfo.event"
        />
            <component :is="Panel"></component>
            <slot></slot>
        </form>
    </div>
</template>