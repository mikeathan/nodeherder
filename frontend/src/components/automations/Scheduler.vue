<script setup lang="ts">
import { ref, watch, PropType, computed } from "vue";
import { getActionType, ActionType, } from "@/contracts/automations"
import { AutomationTriggerAction } from "@/types/automation";
import { emitClosePanel } from "@/mixins/useAutomationsEventBus";
import { PanelComponents } from "@/mixins/usePanelComponents";
import Dropdown from "@/components/controls/Dropdown.vue";
import ButtonPanel from "@/components/controls/ButtonPanel.vue";
import Selection from "@/components/input/Selection.vue";
import { createSaveDeleteButtonItems } from "@/configs/automation/trigger-dropdown.config";

const props = defineProps({
    id: String,
});

const emit = defineEmits<{
    (e: 'save', action: AutomationTriggerAction): void,
    (e: 'delete', action: AutomationTriggerAction): void,
}>()

const buttonPanelItems = computed(() => {
  
  return createSaveDeleteButtonItems(
    () => saveAction(),
    () => removeAction(),
    !actionIsValid,
    !actionIsValid
  );
});

</script>
<template>
    div class="row">
    <ButtonPanel :buttons="buttonPanelItems">
      <Dropdown :items="dropdownItems" class-name="btn-light" :disabled="action.id == ''">
        Add Operation
      </Dropdown>
    </ButtonPanel>
  </div>
    <div>

        ID {{ props.id }}
    </div>
</template>
@