import { KeyValuePair, Nullable } from './types';
export type Automations = Array<Automation>;
export type AutomationMap = KeyValuePair<Automation>;

export type AutomationTriggers = Array<AutomationTrigger>;
export type AutomationTriggerConditions =
  Array<AutomationTriggerCondition>;
export type NumericOperator = '+' | '-' | '*';

export type AutomationActionStep = {
  id: string;
  property: string;
  operator: NumericOperator;
};

export type Automation = {
  id: string;
  friendlyname: string;
  description: string;
  enabled: boolean;
  schedules: TimeSchedule[];
  triggers: AutomationTriggers;
};

export type TimeSchedule = {
  startAt: string;
  name: string;
  type: string;
};

export type AutomationTrigger = {
  name: string;
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
  type: ActionType;
  delay: Nullable<number>;
  steps: Array<AutomationActionStep>;
};
