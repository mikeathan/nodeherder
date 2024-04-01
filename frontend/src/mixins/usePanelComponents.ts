import { defineAsyncComponent } from "vue";

type PanelKey = string;
type Map = { [key: PanelKey]: any };

export const PanelComponents: Map = {
  Trigger: defineAsyncComponent(
    () => import("../components/automations/Trigger.vue")
  ),
  TriggerAction: defineAsyncComponent(
    () => import("../components/automations/actions/TriggerAction.vue")
  ),
  StepAction: defineAsyncComponent(
    () => import("../components/automations/actions/StepAction.vue")
  ),
  ActionEditor: defineAsyncComponent(
    () => import("../components/automations/actions/ActionEditor.vue")
  ),
  ActionViewer: defineAsyncComponent(
    () => import("../components/automations/actions/ActionViewer.vue")
  ),
  PresetRotationAction: defineAsyncComponent(
    () => import("../components/automations/actions/PresetRotationAction.vue")
  ),
};
