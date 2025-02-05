import { ExposeTypes } from '@/types/device.type';
import {
  Automation,
  AutomationTrigger,
  AutomationTriggerCondition,
  AutomationTriggerAction,
  AutomationTriggerConditions,
  AutomationActionStep,
  TimeSchedule,
  AutomationAction,
  AutomationBaseAction,
  AutomationStepAction,
  AutomationPresetCyclingAction,
} from '../types/automation';
import { ValueOf } from '@/types/types.type';

export const EqualityOperators: string[] = ['=', '<=', '>=', '>', '<'];
export const NumericOperators: string[] = ['+', '-', '*'];

export const TimeScheduleTypes: string[] = ['enable', 'disable'];

export type TriggerAction = 'TriggerAction';
export type StepAction = 'StepAction';
export type PresetRotationAction = 'PresetRotationAction';
export type ActionType = TriggerAction | StepAction | PresetRotationAction;

export type TriggerActionOperation = ValueOf<typeof TriggerActionOperations>;
export const TriggerActionOperations = {
  Delay: 'delay',
} as const;

export const AutomationActionTypes = {
  Trigger: 'TriggerAction',
  Step: 'StepAction',
  PresetRotation: 'PresetRotationAction',
} as const;

export class DeviceAutomation implements Automation {
  id: string;
  friendlyname: string;
  description: string;
  enabled: boolean;
  triggers: Array<AutomationTrigger>;
  schedules: TimeSchedule[];

  constructor() {
    this.id = '';
    this.friendlyname = '';
    this.description = '';
    this.enabled = false;
    this.triggers = [];
    this.schedules = [];
  }
}

export class EditableAutomationTrigger implements AutomationTrigger {
  name: string;
  conditions: AutomationTriggerConditions;
  actions: AutomationAction[];

  static create(): AutomationTrigger {
    const trigger = {} as EditableAutomationTrigger;
    trigger.name = '';
    trigger.conditions = [];
    // trigger.action = new EditableActionTrigger(
    //   AutomationActionTypes.Trigger,
    // );

    trigger.actions = [];
    return new EditableAutomationTrigger(trigger);
  }

  static createFrom(trigger: AutomationTrigger): AutomationTrigger {
    return new EditableAutomationTrigger(trigger);
  }

  private constructor(trigger: AutomationTrigger) {
    this.name = trigger.name;
    this.conditions = trigger.conditions;
    this.actions = trigger.actions;
  }
}

export class EditableTriggerCondition implements AutomationTriggerCondition {
  name: string;
  value: any | null;
  equality: string;
  constructor() {
    this.name = '';
    this.value = null;
    this.equality = '=';
  }
}
export function isTriggerAction(action: AutomationAction | null): action is AutomationTriggerAction {
  return action?.type === AutomationActionTypes.Trigger;
}

export function isPresetRotationAction(action: AutomationAction | null): action is AutomationPresetCyclingAction {
  return action?.type === AutomationActionTypes.PresetRotation;
}

export function isStepAction(action: AutomationAction | null): action is AutomationStepAction {
  return action?.type === AutomationActionTypes.Step;
}

export function findActionExposes(action: AutomationAction): string[] {
  if (isStepAction(action)) {
    return action.steps.map((s) => {
      if (s.id == action.id) return s.property;
    }) as string[];
  }

  if (isTriggerAction(action)) {
    return action.exposes.map((expose) => expose.name);
  }
  if (isPresetRotationAction(action)) {
    return [action.property];
  }

  return [];
}
export function createActionFromType(type: ActionType): AutomationAction {
  switch (type) {
    case 'TriggerAction':
      return {
        id: '',
        type: 'TriggerAction',
        exposes: [],
        delay: { unit: 'seconds', value: 0 },
      } as AutomationTriggerAction;
    case 'StepAction':
      return { id: '', type: 'StepAction', steps: [], property: '', data: null } as AutomationStepAction;

    case 'PresetRotationAction':
      return {
        id: '',
        type: 'PresetRotationAction',
        property: '',
      } as AutomationPresetCyclingAction;
  }
}

export function isValid(trigger: AutomationTrigger): boolean {
  const r = trigger.name != '' && trigger.actions.length > 0 && trigger.actions.every((action) => action.id != '');

  return r;
}
