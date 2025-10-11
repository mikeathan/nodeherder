import { KeyValuePair, Nullable, TimeInterval, ValueOf } from './types.type';
export type Automations = Array<Automation>;
export type AutomationMap = KeyValuePair<Automation>;

export type AutomationTriggers = Array<AutomationTrigger>;
export type AutomationTriggerConditions = Array<AutomationCondition>;
export type AutomationActions = Array<AutomationAction>;

export type NumericOperator = '+' | '-' | '*';

// Triggers
export type TriggerType = ValueOf<typeof TriggerTypes>;
export const TriggerTypes = {
  DeviceTrigger: 'deviceTrigger',
  ManualTrigger: 'manualTrigger',
} as const;

// Actions
export type TriggerAction = 'trigger';
export type StepAction = 'step';
export type PresetCyclingAction = 'preset';
export type ActionType = TriggerAction | StepAction | PresetCyclingAction;

export const PublishModes = {
  Batch: 'batch',
  Single: 'single',
} as const;
export type PublishMode = ValueOf<typeof PublishModes>;
export const AutomationActionTypes = {
  Trigger: 'trigger',
  Step: 'step',
  PresetCycling: 'preset',
} as const;

// Operations
export type TriggerActionOperation = ValueOf<typeof TriggerActionOperations>;
export const TriggerActionOperations = {
  Delay: 'delay',
} as const;

export type ExposeConditionType = 'expose';
export type TimeConditionType = 'time';
export type ConditionType = ExposeConditionType | TimeConditionType;

export const AutomationConditionTypes = {
  Expose: 'expose',
  Time: 'time',
} as const;

// Automations
export type Automation = {
  id: string;
  friendlyname: string;
  type: string;
  description: string;
  enabled: boolean;
  schedules: TimeSchedule[];
  triggers: AutomationTriggers;
};

export type TimeRange = {
  startAt: string;
  endAt: string;
};

export type TimeScheduleType = 'enable' | 'disable';

export type TimeSchedule = {
  startAt: string;
  type: TimeScheduleType;
};

export type AutomationActionStep = {
  id: string;
  property: string;
  operator: NumericOperator;
};

export type AutomationTrigger = {
  name: string;
  type: TriggerType;
  conditions: AutomationTriggerConditions;
  actions: AutomationActions;
};

export type ExposeCondition = {
  type: ConditionType;
  name: string;
  value: Nullable<any>;
  equality: string;
};

export type TimeCondition = {
  type: ConditionType;
  timeRange: TimeRange;
};

export const isExposeCondition = (condition: AutomationCondition): condition is ExposeCondition => {
  return condition.type === 'expose';
};

export const isTimeCondition = (condition: AutomationCondition): condition is TimeCondition => {
  return condition.type === 'time';
};

export type AutomationCondition = ExposeCondition | TimeCondition;

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
  delay?: TimeInterval | undefined;
  publishMode: PublishMode;
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
