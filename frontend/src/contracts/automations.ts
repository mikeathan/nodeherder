import { ExposeTypes } from '@/types/device.type';
import {
  Automation,
  AutomationTrigger,
  AutomationTriggerCondition,
  AutomationTriggerAction,
  AutomationTriggerConditions,
  AutomationActionStep,
  TimeSchedule,
} from '../types/automation';
import { ExposeType } from '../types/device';
import { ValueOf } from '@/types/types';

export const EqualityOperators: string[] = [
  '=',
  '<=',
  '>=',
  '>',
  '<',
];
export const NumericOperators: string[] = ['+', '-', '*'];

export const TimeScheduleTypes: string[] = [
  'enable',
  'disable',
];

export type TriggerAction = 'TriggerAction';
export type StepAction = 'StepAction';
export type PresetRotationAction = 'PresetRotationAction';
export type ActionType =
  | TriggerAction
  | StepAction
  | PresetRotationAction;

export type TriggerActionOperation = ValueOf<
  typeof TriggerActionOperations
>;
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

export class EditableAutomationTrigger
  implements AutomationTrigger
{
  name: string;
  conditions: AutomationTriggerConditions;
  actions: AutomationTriggerAction[];

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

  static createFrom(
    trigger: AutomationTrigger,
  ): AutomationTrigger {
    return new EditableAutomationTrigger(trigger);
  }

  private constructor(trigger: AutomationTrigger) {
    this.name = trigger.name;
    this.conditions = trigger.conditions;
    this.actions = trigger.actions;
  }
}

export class EditableTriggerCondition
  implements AutomationTriggerCondition
{
  name: string;
  value: any | null;
  equality: string;
  constructor() {
    this.name = '';
    this.value = null;
    this.equality = '=';
  }
}

export class EditableActionTrigger
  implements AutomationTriggerAction
{
  id: string;
  friendlyname: string;
  property: string;
  data: any | null;
  delay: number | null;
  steps: AutomationActionStep[];
  type: ActionType;

  constructor(type: ActionType) {
    this.id = '';
    this.friendlyname = '';
    this.property = '';
    this.data = null;
    this.delay = null;
    this.type = type;
    this.steps = new Array<AutomationActionStep>();
  }

  public setProperty(value: string): void {
    this.property = value;
    this.delay = null;
    this.data = null;
  }

  public setDeviceId(
    id: string,
    friendlyname: string,
  ): void {
    this.id = id;
    this.friendlyname = friendlyname;
  }
}

export function getActionType(
  action: AutomationTriggerAction,
): ActionType {
  const editableAction = action as EditableActionTrigger;
  if (editableAction.type != undefined) {
    return editableAction.type;
  }
  return action.steps.length > 0
    ? AutomationActionTypes.Step
    : AutomationActionTypes.Trigger;
}

export function clearAction(
  action: AutomationTriggerAction,
): void {
  action.id = '';
  action.friendlyname = '';
  action.property = '';
  action.data = null;
  action.delay = null;
}

export function insertTriggerCondition(
  trigger: AutomationTrigger,
  newCondition?: AutomationTriggerCondition,
) {
  trigger.conditions.push(
    newCondition ?? new EditableTriggerCondition(),
  );
}

export function insertCondition(
  conditions: AutomationTriggerConditions,
  newCondition?: AutomationTriggerCondition,
) {
  conditions.push(
    newCondition ?? new EditableTriggerCondition(),
  );
}
export function removeTriggerCondition(
  trigger: AutomationTrigger,
  condition: AutomationTriggerCondition,
) {
  trigger.conditions = trigger.conditions.filter(
    (c) => c != condition,
  );
}

export function removeCondition(
  conditions: AutomationTriggerConditions,
  condition: AutomationTriggerCondition,
) {
  conditions = conditions.filter((c) => c != condition);
}

export function isValid(
  trigger: AutomationTrigger,
): boolean {
  const r =
    trigger.name != '' &&
    trigger.actions.length > 0 &&
    trigger.actions.every(
      (action) => action.id != '' && action.property != '',
    );

  return r;
}

export function setDeviceId(
  action: AutomationTriggerAction,
  id: string,
  friendlyname: string,
): void {
  action.id = id;
  action.friendlyname = friendlyname;
}

export function setProperty(
  action: AutomationTriggerAction,
  value: string,
  type: ExposeType,
): void {
  action.property = value;

  // reset remaining properties
  action.delay = null;

  if (
    type == ExposeTypes.Binary ||
    type == ExposeTypes.Enum
  ) {
    action.data = '';
  } else {
    action.data = 0;
  }
}
