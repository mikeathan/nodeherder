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

// export interface  TriggerWrapper{

//   hasConditions(): boolean;
//   addCondition(condition: Condition) :void;

//   removeLastCondition(): void ;

//   remove(condition: Condition): void ;
//   idx(idx: number): void ;
//    friendlyName(name: string): void;
//    displayName(): string ;

//    setAction(action: Action): void;

// }

export class ExposeTriggerWrapper {
  trigger: ExposeTrigger;
  //private _action: ActionTriggerWrapper;

  constructor(trigger: ExposeTrigger) {
    this.trigger = trigger;
    //this._action = new ActionTriggerWrapper(trigger.action);
    this.trigger.conditions.forEach(function callback(condition, index) {
      condition.idx = index + 1;
    });
  }

  getTrigger(): ExposeTrigger {
    return this.trigger;
  }
  isValid(): boolean {
    return this.trigger.name != "" && this.trigger.action != null;
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

  public action(): Action | null {
    return this.trigger.action;
  }

  public setAction(value: Action): void {
    this.trigger.action = value;
  }

  public clearAction(): void {
    this.trigger.action = null;
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
  action: ActionTrigger | null;

  constructor(name: string) {
    this.idx = -1;
    this.name = name;
    this.conditions = [];
    this.action = null;
  }
}

abstract class Action {
  id: string;
  friendlyname: string;
  property: string;
  data: any;
  delay: number | null;
  step: number;

  constructor() {
    this.id = "";
    this.friendlyname = "";
    this.property = "";
    this.data = null;
    this.delay = null;
    this.step = 0;
  }
}

export class ActionTriggerWrapper {
  _action: Action | null;
  // id: string;
  // friendlyname: string;
  // property: string;
  // data: any;
  // delay: number | null;
  // step: number;

  constructor(action: Action | null) {
    this._action = action;
    // this.id = "";
    // this.friendlyname = "";
    // this.property = "";
    // this.data = null;
    // this.delay = null;
    // this.step = 0;
  }

  // public set friendlyname(value: string): void {
  //   this._action.friendlyname = value;
  // }
  // public get friendlyname(): string {
  //   return this._action.friendlyname;
  // }

  public create(): void {
    this._action = new ActionTrigger();
  }

  public clear(): void {
    this._action = null;
  }
}

export class ActionTrigger extends Action {
  constructor() {
    super();
  }
}

export const EqualityOperators: string[] = ["=", "<=", ">=", ">", "<"];

interface Operation {
  name: string;
  value: number;
}

export const StepOperators: { [Name: string]: number } = {
  None: 0,
  increase: 1,
  decrease: 2,
};

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
  //return value.charAt(0).toUpperCase() + value.slice(1);
}
