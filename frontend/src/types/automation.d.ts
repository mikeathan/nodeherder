import { TimeInterval } from './settings.type';
import { KeyValuePair, Nullable, ValueOf } from './types.type';
export type Automations = Array<Automation>;
export type AutomationMap = KeyValuePair<Automation>;

export type AutomationTriggers = Array<AutomationTrigger>;
export type AutomationTriggerConditions = Array<AutomationTriggerCondition>;
export type AutomationActions = Array<AutomationAction>;

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

export type TimeScheduleType = 'enable' | 'disable';

export type TimeSchedule = {
  startAt: string;
  type: TimeScheduleType;
};

export type AutomationTrigger = {
  name: string;
  conditions: AutomationTriggerConditions;
  actions: AutomationActions;
};

export type AutomationTriggerCondition = {
  name: string;
  value: Nullable<any>;
  equality: string;
};

type AutomationBaseAction = {
  id: string;
  type: ActionType;
};

export type AutomationTriggerActionExpose = {
  name: string;
  data: any;
};

export type AutomationTriggerAction = AutomationBaseAction & {
  exposes: Array<AutomationTriggerActionExpose>;
  delay: TimeInterval;
};

export type AutomationStepAction = AutomationBaseAction & {
  property: string;
  steps: Array<AutomationActionStep>;
  data: any;
};

export type AutomationPresetCyclingAction = AutomationBaseAction & {
  property: string;
};

export type AutomationAction = AutomationTriggerAction | AutomationStepAction | AutomationPresetCyclingAction;
