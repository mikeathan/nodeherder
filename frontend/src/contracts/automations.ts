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
} from '../types/automation';
import { ExposeType } from '../types/device';
import { ValueOf } from '@/types/types.type';
import { TimeInterval } from '@/types/settings.type';

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

export class EditableActionTrigger implements AutomationBaseAction {
  id: string;
  friendlyname: string;
  property: string;
  data: any | null;
  delay: TimeInterval;
  steps: AutomationActionStep[];
  type: ActionType;

  constructor(type: ActionType) {
    this.id = '';
    this.friendlyname = '';
    this.property = '';
    this.data = null;
    this.delay = {
      value: 0,
      unit: 'seconds',
    };
    this.type = type;
    this.steps = new Array<AutomationActionStep>();
  }

  static createTriggerAction(): EditableActionTrigger {
    return new EditableActionTrigger('TriggerAction');
  }

  static createStepAction(): EditableActionTrigger {
    return new EditableActionTrigger('StepAction');
  }

  static createPresetAction(): EditableActionTrigger {
    return new EditableActionTrigger('PresetRotationAction');
  }

  toMqttAction(): AutomationAction {
    switch (this.type) {
      case 'TriggerAction':
        return {
          id: this.id,
          property: this.property,
          type: this.type,
          exposes: [{ name: this.property, data: this.data }],
          delay: this.delay,
        };
      case 'StepAction':
        return {
          id: this.id,
          property: this.property,
          type: this.type,
          steps: this.steps,
          data: this.data,
        };
      case 'PresetRotationAction':
        return {
          id: this.id,
          property: this.property,
          type: this.type,
        };
    }
  }
}

export function getActionType(action: AutomationAction): ActionType {
  const editableAction = action as EditableActionTrigger;
  if (editableAction.type != undefined) {
    return editableAction.type;
  }
  return action.steps.length > 0 ? AutomationActionTypes.Step : AutomationActionTypes.Trigger;
}

export function isValid(trigger: AutomationTrigger): boolean {
  const r =
    trigger.name != '' &&
    trigger.actions.length > 0 &&
    trigger.actions.every((action) => action.id != '' && action.property != '');

  return r;
}
