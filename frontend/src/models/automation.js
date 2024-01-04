export class DeviceTrigger {
  constructor() {
    this.id = "";
    this.friendlyname = "";
    this.description = "";
    this.enabled = false;
    this.triggers = [];
  }

  toJson() {
    return JSON.stringify(this, converter);
  }
}

export class ActionTrigger {
  constructor() {
    this.id = "";
    this.friendlyname = null;
    this.type = null;
    this.property = null;
    this.data = null;
    this.delay = null;
    this.step = null;
  }
}

export class ExposeTrigger {
  constructor(name) {
    this.idx = -1;
    this.name = name;
    this.conditions = [];
    this.action = null;
  }

  hasConditions() {
    return this.conditions.length != 0;
  }

  addCondition(condition) {
    this.conditions.push(condition);
  }

  removeLastCondition() {
    if (this.conditions.length >= 0) {
      this.conditions = this.conditions.slice(0, -1);
    }
  }

  remove(condition) {
    let index = this.Conditions.indexOf(condition);
    if (index !== -1) {
      this.conditions.splice(index, 1);
    }
  }
}

export class Condition {
  constructor() {
    this.name = "";
    this.equality = "=";
    this.value = "";
    this.idx = 0;
  }
}

function converter(key, value) {
  if (key == "idx") return undefined;
  if (value == null) return undefined;

  return value;
}
