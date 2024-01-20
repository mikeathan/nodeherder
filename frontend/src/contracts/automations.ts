import {
  Automation,
  AutomationTrigger,
  AutomationTriggerCondition,
  AutomationTriggerAction,
} from "../types/automation";

export interface DeviceAutomation extends Automation {}
export interface DeviceAutomationTrigger extends AutomationTrigger {}
export interface DeviceAutomationCondition extends AutomationTriggerCondition {}
export interface DeviceAutomationAction extends AutomationTriggerAction {}

export class DeviceTriggerClass implements DeviceAutomation {
  id: string;
  friendlyname: string;
  description: string;
  enabled: boolean;
  triggers: Array<DeviceAutomationTrigger>;

  constructor() {
    this.id = "";
    this.friendlyname = "";
    this.description = "";
    this.enabled = false;
    this.triggers = [];
  }
}

// #########################################
export class DeviceTrigger {
  id: string;
  friendlyname: string;
  description: string;
  enabled: boolean;
  triggers: Array<ExposeTrigger>;

  constructor() {
    this.id = "";
    this.friendlyname = "";
    this.description = "";
    this.enabled = false;
    this.triggers = [];
  }
}

export class ExposeTriggerWrapper {
  trigger: ExposeTrigger;

  constructor(trigger: ExposeTrigger) {
    this.trigger = trigger;
    this.trigger.conditions.forEach(function callback(condition, index) {
      condition.idx = index + 1;
    });
  }

  getTrigger(): ExposeTrigger {
    return this.trigger;
  }

  isValid(): boolean {
    return this.trigger.name != "" && this.trigger.action.id != "";
  }

  getConditions(): Array<Condition> {
    return this.trigger.conditions;
  }

  hasConditions(): boolean {
    return this.trigger.conditions.length != 0;
  }

  addCondition(condition: Condition) {
    this.trigger.conditions.push(condition);
  }

  removeLastCondition(): void {
    if (this.trigger.conditions.length >= 0) {
      this.trigger.conditions = this.trigger.conditions.slice(0, -1);
    }
  }

  removeCondition(index: number): void {
    this.trigger.conditions = this.trigger.conditions.filter(
      (k) => k.idx != index
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

  public name(): string {
    return this.trigger.name;
  }

  public displayName(): string {
    return capitalizeText(this.trigger.name);
  }

  public action(): Action {
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
    this.trigger.action = new ActionTrigger();
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

export class DefaultExposeTriggerWrapper extends ExposeTriggerWrapper {
  constructor() {
    super(new ExposeTrigger(""));
  }
}

export class ExposeTrigger {
  idx: number;
  name: string;
  conditions: Array<Condition>;
  action: ActionTrigger;

  constructor(name: string) {
    this.idx = -1;
    this.name = name;
    this.conditions = [];
    this.action = new ActionTrigger();
  }
}

export abstract class Action {
  id: string;
  friendlyname: string;
  property: string;
  data: any;
  delay: number | null;
  operation: number;

  constructor() {
    this.id = "";
    this.friendlyname = "";
    this.property = "";
    this.data = null;
    this.delay = null;
    this.operation = 0;
  }

  public abstract setProperty(value: string): void;
  public abstract setDeviceId(id: string, friendlyname: string): void;
}

export class ActionTrigger extends Action {
  constructor() {
    super();
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

export class Condition {
  name: string;
  equality: string;
  value: any | null;
  idx: number;

  constructor() {
    this.name = "";
    this.equality = EqualityOperators[0];
    this.value = "";
    this.idx = 0;
  }
}

// todo: convert to extension class
function capitalizeText(value: string): string {
  return value
    .toLowerCase()
    .split(" ")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}
