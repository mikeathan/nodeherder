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

abstract class Action {
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
}

export class ActionTrigger extends Action {
  constructor() {
    super();
  }
}

export const EqualityOperators: string[] = ["=", "<=", ">=", ">", "<"];

export interface operationItem {
  name: string;
  value: number;
}


const ActionOperationsMap: { [K in string]: operationItem } = {
  "numeric": { name: "Steps", value: -1 },
  "presets": { name: "Rotation", value: 3 },
  "step_increase": { name: "Increase", value: 1 },
  "step_decrease": { name: "Decrease", value: 2 },
} as const;

export function resolveStepOperations() {
  let arr = Array<operationItem>(ActionOperationsMap.step_increase, ActionOperationsMap.step_decrease)
  var r = Object.assign({}, ...arr.map((x) => ({ [x.name]: x.value })));

  return r
}

export function resolveObjectOperations(obj: any) {
  let arr = Array<operationItem>({ name: "None", value: 0 })
  if (obj.hasOwnProperty("type")) {
    if (obj["type"] === "numeric") {
      arr.push(ActionOperationsMap.numeric)
    }
  }
  if (obj.hasOwnProperty("presets")) {
    if (obj["presets"] !== undefined) {
      arr.push(ActionOperationsMap.presets)
    }
  }

  var r = Object.assign({}, ...arr.map((x) => ({ [x.name]: x.value })));

  return r
}



export type OperationTypeKeys = keyof typeof OperationType;
export enum OperationType {
  NoOperation = 0,
  StepOperation = -1,
  RotationOperation = 3,
  StepIncreaseOperation = 1,
  StepDecreaseOperation = 2,
}


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
