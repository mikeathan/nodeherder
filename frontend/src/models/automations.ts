class DeviceTrigger {
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

class ActionTrigger {
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

class ExposeTrigger {
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

  hasConditions(): boolean {
    return this.conditions.length != 0;
  }

  addCondition(condition: Condition) {
    this.conditions.push(condition);
  }

  removeLastCondition(): void {
    if (this.conditions.length >= 0) {
      this.conditions = this.conditions.slice(0, -1);
    }
  }

  remove(condition: Condition): void {
    let index = this.conditions.indexOf(condition);
    if (index !== -1) {
      this.conditions.splice(index, 1);
    }
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

class Condition {
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
