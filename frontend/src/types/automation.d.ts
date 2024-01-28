import { KeyyValuePair, Nullable } from "./types";
export type Automations = Array<Automation>;
export type AutomationMap = KeyyValuePair<Automation>;

export type AutomationTriggers = Array<AutomationTrigger>;
export type AutomationTriggerConditions = Array<AutomationTriggerCondition>;

export type Automation = {
  id: string;
  friendlyname: string;
  description: string;
  enabled: boolean;
  triggers: AutomationTriggers;
};

export interface AutomationTriggerWrapper extends AutomationTrigger {
  name: string;
  idx: number; // TEMPORARY - need to remove!!!
  conditions: AutomationTriggerConditions;
  action: AutomationTriggerAction;
  isValid(): boolean;

  addCondition();

  removeCondition(condition: AutomationTriggerCondition): void;

  displayName(): string;

  setActionDeviceId(id: string, friendlyname: string): void;

  setActionProperty(value: string): void;

  clearAction(): void;
}

export type AutomationTrigger = {
  name: string;
  idx: number; // TEMPORARY - need to remove!!!
  conditions: AutomationTriggerConditions;
  action: AutomationTriggerAction;
};

export type AutomationTriggerCondition = {
  name: string;
  value: Nullable<any>;
  equality: string;
};

export type AutomationTriggerAction = {
  id: string;
  friendlyname: string;
  property: string;
  data: Nullable<any>;
  operation: number;
  delay: Nullable<number>;
};
