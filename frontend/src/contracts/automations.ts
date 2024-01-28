import {
  Automation,
  AutomationTrigger,
  AutomationTriggerCondition,
  AutomationTriggerAction,
  AutomationTriggerConditions,
} from "../types/automation";
import { Nullable } from "../types/types";
import { capitalizeText } from "../modules/formatters/text.formatter";

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
  idx: number; // TEMPORARY - need to remove!!!
  conditions: AutomationTriggerConditions;
  action: AutomationTriggerAction;

  static create(): EditableAutomationTrigger {
    const trigger = {} as AutomationTrigger;
    trigger.name = "";
    trigger.idx = -1;
    trigger.conditions = [];
    trigger.action = new EditableActionTrigger();
    return new EditableAutomationTrigger(trigger);
  }

  static createFrom(trigger: AutomationTrigger): EditableAutomationTrigger {
    return new EditableAutomationTrigger(trigger);
  }

  private constructor(trigger: AutomationTrigger) {
    this.name = trigger.name;
    this.idx = trigger.idx;
    this.conditions = trigger.conditions;
    this.action = trigger.action;
  }

  isValid(): boolean {
    return (
      this.name != "" && this.action.id != "" && this.action.property != ""
    );
  }

  hasConditions(): boolean {
    return this.conditions.length != 0;
  }

  addCondition() {
    this.conditions.push(new EditableTriggerCondition());
  }

  removeCondition(condition: AutomationTriggerCondition): void {
    this.conditions = this.conditions.filter((c) => c != condition);
  }

  public displayName(): string {
    return capitalizeText(this.name);
  }

  public setActionDeviceId(id: string, friendlyname: string): void {
    this.action.id = id;
    this.action.friendlyname = friendlyname;
  }

  public setActionProperty(value: string): void {
    this.action.property = value;

    // reset value
    this.action.operation = 0;
    this.action.delay = null;
    this.action.data = null;
  }

  public clearAction(): void {
    this.action.id = "";
    this.action.friendlyname = "";
    this.action.property = "";
    this.action.data = null;
    this.action.operation = 0;
    this.action.delay = null;
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
