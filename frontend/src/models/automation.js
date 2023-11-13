export const Operators = [
  { text: "=", value: "=" },
  { text: "<=", value: "<=" },
  { text: ">=", value: ">=" },
  { text: ">", value: ">" },
  { text: "<", value: "<" },
];
export class DeviceTrigger {
  constructor() {
    this.id = "";
    this.friendly_name = "";
    this.description = "";
    this.enabled = false;
    this.triggers = [];
  }

  // getProperties(exposeName) {
  //   var expose = this.device.exposes[exposeName];
  //   if (expose == undefined || expose.properties == undefined) {
  //     return [];
  //   }
  //   return expose.properties;
  // }
  // getAttributes(exposeName) {
  //   var expose = this.device.exposes[exposeName];
  //   if (expose == undefined || expose.attributes == undefined) {
  //     return [];
  //   }
  //   return expose.attributes;
  // }
}

export class ExposeTrigger {
  constructor(name) {
    this.name = name;
    this.conditions = [];
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
      this.Conditions.splice(index, 1);
    }
  }
}

export class Condition {
  constructor(name, operator, data) {
    this.name = name;
    this.operator = operator;
    this.data = data;
    this.id = 0;
  }
}
