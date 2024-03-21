import { defineAsyncComponent } from "vue";

export const PanelComponents = {
  TriggerAction: defineAsyncComponent(
    () => import("../components/automations/actions/TriggerAction.vue")
  ),
  StepAction: defineAsyncComponent(
    () => import("../components/automations/actions/StepAction.vue")
  ),
};
