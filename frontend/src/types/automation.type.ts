import { KeyValuePair, Nullable, TimeInterval, ValueOf } from './types.type';
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

export const PublishModes = ['batch', 'single'] as const;
export type PublishMode = ValueOf<typeof PublishModes>;
export const AutomationActionTypes = {
  Trigger: 'trigger',
  Step: 'step',
  PresetCycling: 'preset',
} as const;

export type TriggerActionOperation = ValueOf<typeof TriggerActionOperations>;
export const TriggerActionOperations = {
  Delay: 'delay',
} as const;
export type ExposeConditionType = 'expose';

export type ConditionType = ExposeConditionType;
export const AutomationConditionTypes = {
  Expose: 'expose',
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

export type TimeRange = {
  startAt: string;
  endAt: string;
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

export type ExposeCondition = {
  type: ConditionType;
  name: string;
  value: Nullable<any>;
  equality: string;
  timeRange?: TimeRange;
};

export type AutomationCondition = ExposeCondition;

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
