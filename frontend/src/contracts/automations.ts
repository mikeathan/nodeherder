import { ExposeTypes } from "@/types/device.type";
import {
  Automation,
  AutomationTrigger,
  AutomationTriggerCondition,
  AutomationTriggerAction,
  AutomationTriggerConditions,
} from "../types/automation";
import { ExposeType } from "../types/device";
//import { Nullable } from "../types/types";

export const EqualityOperators: string[] = ["=", "<=", ">=", ">", "<"];

export class DeviceAutomation implements Automation {
  id: string;
  friendlyname: string;
  description: string;
  enabled: boolean;
  triggers: Array<AutomationTrigger>;

  constructor() {
    this.id = "";
    this.friendlyname = "";
    this.description = "";
    this.enabled = false;
    this.triggers = [];
  }
}

export class EditableAutomationTrigger implements AutomationTrigger {
  name: string;
  conditions: AutomationTriggerConditions;
  action: AutomationTriggerAction;

  static create(): AutomationTrigger {
    const trigger = {} as EditableAutomationTrigger;
    trigger.name = "";
    trigger.conditions = [];
    trigger.action = new EditableActionTrigger();

    return new EditableAutomationTrigger(trigger);
  }

  static createFrom(trigger: AutomationTrigger): AutomationTrigger {
    return new EditableAutomationTrigger(trigger);
  }

  private constructor(trigger: AutomationTrigger) {
    this.name = trigger.name;
    this.conditions = trigger.conditions;
    this.action = trigger.action;
  }
}

export class EditableTriggerCondition implements AutomationTriggerCondition {
  name: string;
  value: any | null;
  equality: string;
  constructor() {
    this.name = "";
    this.value = null;
    this.equality = "=";
  }
}

export class EditableActionTrigger implements AutomationTriggerAction {
  id: string;
  friendlyname: string;
  property: string;
  data: any | null;
  operation: number;
  delay: number | null;

  constructor() {
    this.id = "";
    this.friendlyname = "";
    this.property = "";
    this.data = null;
    this.operation = 0;
    this.delay = null;
  }

  public setProperty(value: string): void {
    this.property = value;
    this.operation = 0;
    this.delay = null;
    this.data = null;
  }

  public setDeviceId(id: string, friendlyname: string): void {
    this.id = id;
    this.friendlyname = friendlyname;
  }
}

export function clearAction(action: AutomationTriggerAction): void {

  action.id = "";
  action.friendlyname = "";
  action.property = "";
  action.data = null;
  action.operation = 0;
  action.delay = null;
}

export function insertCondition(
  trigger: AutomationTrigger,
  newCondition?: AutomationTriggerCondition
) {
  trigger.conditions.push(newCondition ?? new EditableTriggerCondition());
}

export function removeCondition(
  trigger: AutomationTrigger,
  condition: AutomationTriggerCondition
) {
  trigger.conditions = trigger.conditions.filter((c) => c != condition);
}

export function isValid(automation: AutomationTrigger): boolean {
  return (
    automation.name != "" &&
    automation.action.id != "" &&
    automation.action.property != ""
  );
}

export function setDeviceId(
  action: AutomationTriggerAction,
  id: string,
  friendlyname: string
): void {
  action.id = id;
  action.friendlyname = friendlyname;
}

export function setProperty(
  action: AutomationTriggerAction,
  value: string,
  type: ExposeType
): void {
  action.property = value;

  // reset remaining properties
  action.operation = 0;
  action.delay = null;

  if (type == ExposeTypes.Binary || type == ExposeTypes.Enum) {
    action.data = "";
  } else {
    action.data = 0;
  }
}
