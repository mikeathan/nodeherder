import {
  Automation,
  AutomationTrigger,
  AutomationCondition,
  AutomationTriggerAction,
  AutomationTriggerConditions,
  TimeSchedule,
  AutomationAction,
  AutomationStepAction,
  AutomationPresetCyclingAction,
  AutomationActionTypes,
  ActionType,
  ConditionType,
  AutomationConditionTypes,
} from '../types/automation.type.js';

export const EqualityOperators: string[] = ['=', '<=', '>=', '>', '<'];
export const NumericOperators: string[] = ['+', '-', '*'];

export const TimeScheduleTypes: string[] = ['enable', 'disable'];

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

export function isTriggerAction(action: AutomationAction | null): action is AutomationTriggerAction {
  return action?.type === AutomationActionTypes.Trigger;
}

export function isPresetCyclingAction(action: AutomationAction | null): action is AutomationPresetCyclingAction {
  return action?.type === AutomationActionTypes.PresetCycling;
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
  if (isPresetCyclingAction(action)) {
    return [action.property];
  }

  return [];
}
export function createConditionFromType(type: ConditionType): AutomationCondition {
  switch (type) {
    case AutomationConditionTypes.Expose:
      return {
        type: type,
        name: '',
        value: null,
        equality: '=',
      } as AutomationCondition;
  }
}

export function createActionFromType(type: ActionType): AutomationAction {
  switch (type) {
    case AutomationActionTypes.Trigger:
      return {
        id: '',
        type: type,
        exposes: [],
        splitCommands: false,
        delay: { unit: 'seconds', value: 0 },
      } as AutomationTriggerAction;
    case AutomationActionTypes.Step:
      return { id: '', type: type, steps: [], property: '', data: null } as AutomationStepAction;

    case AutomationActionTypes.PresetCycling:
      return {
        id: '',
        type: type,
        property: '',
      } as AutomationPresetCyclingAction;
  }
}

export function isValid(trigger: AutomationTrigger): boolean {
  const r = trigger.name != '' && trigger.actions.length > 0 && trigger.actions.every((action) => action.id != '');

  return r;
}
