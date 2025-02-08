import { TimeInterval } from './settings.type';
import { KeyValuePair, Nullable, ValueOf } from './types.type';
export type Automations = Array<Automation>;
export type AutomationMap = KeyValuePair<Automation>;

export type AutomationTriggers = Array<AutomationTrigger>;
export type AutomationTriggerConditions = Array<AutomationCondition>;
export type AutomationActions = Array<AutomationAction>;

export type NumericOperator = '+' | '-' | '*';

export type TriggerAction = 'trigger';
export type StepAction = 'step';
export type PresetCyclingAction = 'preset';
export type ActionType = TriggerAction | StepAction | PresetCyclingAction;
export const AutomationActionTypes = {
  Trigger: 'trigger',
  Step: 'step',
  PresetCycling: 'preset',
} as const;

export type ExposeConditionType = 'expose';
export type TimeCondition = 'time';
export type ConditionType = ExposeConditionType | TimeCondition;
export const AutomationConditionTypes = {
  Expose: 'expose',
  Time: 'time',
} as const;

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

export type AutomationCondition = {
  type: ConditionType;
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
