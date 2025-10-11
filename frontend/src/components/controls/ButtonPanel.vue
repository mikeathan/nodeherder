<script setup lang="ts">
  import { PropType } from 'vue';
  import { ButtonPanelType, ButtonType, DropDownType, DropDownItemType, Severity, Size } from '@/types/controls.type';

  const props = defineProps({
    buttons: {
      type: Object as PropType<Array<ButtonPanelType>>,
      default: [],
      required: true,
    },
    severity: {
      type: String as PropType<Severity>,
      default: 'primary',
    },
    size: {
      type: String as PropType<Size>,
    }
  });

  function isButtonType(item: ButtonPanelType): item is ButtonType {
    return item.hasOwnProperty('label') && !item.hasOwnProperty('items');
  }
  function isDropdownType(item: ButtonPanelType): item is DropDownType {
    return item.hasOwnProperty('label') && item.hasOwnProperty('items');
  }

  function createEvent(event: Event, button: ButtonType) {
    event.preventDefault();
    try {
      button.command(event);
    } catch (error) {
      console.error('Error during button click:', error);
    }
  }

  function createDropEvent(event: Event, button: DropDownItemType) {
    event.preventDefault();
    try {
      button.command(event);
    } catch (error) {
      console.error('Error during button click:', error);
    }
  }
</script>

<!-- // if button is dropdown use <SplitButton label="Save" @click="save" :model="items" /> -->
<template>
  <div class="grid grid-cols-4 gap-1">
    <div v-for="item in props.buttons" :key="item.label">
      <template v-if="isButtonType(item)">
        <Button
          :key="item.label"
          :label="item.label"
          :severity="props.severity"
          variant="outlined"
          :disabled="item.disabled"
          @click="(event) => createEvent(event, item)"
          raised />
      </template>
      <template v-else-if="isDropdownType(item)">
        <SplitButton
          v-for="dropdownItem in item.items"
          :key="dropdownItem.label"
          :label="dropdownItem.label"
          @click="(event) => createDropEvent(event, dropdownItem)" />
      </template>
    </div>
    <!-- <Button
        v-for="buttonInfo in props.buttons"
        :key="buttonInfo.name"
        :label="buttonInfo.name"
        :disabled="buttonInfo.disabled"
        @click="(event)=>createEvent(event, buttonInfo as ButtonType)" /> -->
    <!-- <component :is="Panel"></component> -->
    <!-- <slot></slot> -->
  </div>
</template>
