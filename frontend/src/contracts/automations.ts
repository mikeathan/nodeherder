import {
  Automation,
  AutomationTrigger,
  AutomationTriggerCondition,
  AutomationTriggerAction,
  AutomationTriggerConditions,
} from "../types/automation";
import { Nullable } from "../types/types";

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
  trigger: AutomationTrigger;

  name: string;
  idx: number; // TEMPORARY - need to remove!!!
  conditions: AutomationTriggerConditions;
  action: AutomationTriggerAction;

  static create(): EditableAutomationTrigger {
    const trigger = {} as AutomationTrigger;
    trigger.name = "";
    trigger.idx = -1;
    trigger.conditions = [];
    trigger.action = {} as AutomationTriggerAction;
    return new EditableAutomationTrigger(trigger);
  }

  static createFrom(trigger: AutomationTrigger): EditableAutomationTrigger {
    return new EditableAutomationTrigger(trigger);
  }

  private constructor(trigger: AutomationTrigger) {
    this.trigger = trigger;

    this.name = trigger.name;
    this.idx = trigger.idx;
    this.conditions = trigger.conditions;
    this.action = trigger.action;
  }

  getTrigger(): AutomationTrigger {
    return this.trigger;
  }

  isValid(): boolean {
    return this.trigger.name != "" && this.trigger.action.id != "" && this.trigger.action.property != "";
  }

  getConditions(): Array<AutomationTriggerCondition> {
    return this.trigger.conditions;
  }

  hasConditions(): boolean {
    return this.trigger.conditions.length != 0;
  }

  addCondition() {
    this.trigger.conditions.push(new EditableTriggerCondition());
  }

  removeConditionByValue(condition: AutomationTriggerCondition): void {
    this.trigger.conditions = this.trigger.conditions
      .filter(
        (c) => c != condition
      );
  }

  public setIdx(idx: number): void {
    this.trigger.idx = idx;
  }

  public getIdx(): number {
    return this.trigger.idx;
  }

  public setName(name: string): void {
    this.trigger.name = name;
  }

  public getName(): string {
    return this.trigger.name;
  }

  public displayName(): string {
    return capitalizeText(this.trigger.name);
  }

  public getAction(): AutomationTriggerAction {
    return this.trigger.action;
  }

  public setActionDeviceId(id: string, friendlyname: string): void {
    this.trigger.action.id = id;
    this.trigger.action.friendlyname = friendlyname;
  }

  public setActionProperty(value: string): void {
    this.trigger.action.property = value;
    this.trigger.action.operation = 0;
    this.trigger.action.delay = null;
    this.trigger.action.data = null;
  }

  public createAction(): void {
    this.trigger.action = {} as AutomationTriggerAction;
  }

  public clearAction(): void {
    this.trigger.action.id = "";
    this.trigger.action.friendlyname = "";
    this.trigger.action.property = "";
    this.trigger.action.data = null;
    this.trigger.action.operation = 0;
    this.trigger.action.delay = null;
  }
}

export class EditableTriggerCondition implements AutomationTriggerCondition {
  name: string;
  value: Nullable<any>;
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
  data: Nullable<any>;
  operation: number;
  delay: Nullable<number>;

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

export const EqualityOperators: string[] = ["=", "<=", ">=", ">", "<"];

// todo: convert to extension class
function capitalizeText(value: string): string {
  return value
    .toLowerCase()
    .split(" ")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}
