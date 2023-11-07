export const Operators = [
  { text: "=", value: "=" },
  { text: "<=", value: "<=" },
  { text: ">=", value: ">=" },
  { text: ">", value: ">" },
  { text: "<", value: "<" },
];
export class DeviceTrigger {
  constructor(id) {
    this.id = id;
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
    this.Name = name;
    this.Conditions = [];
  }

  hasConditions() {
    return this.Conditions.length != 0;
  }

  addCondition(condition) {
    this.Conditions.push(condition);
  }

  removeLastCondition() {
    if (this.Conditions.length >= 0) {
      this.Conditions = this.Conditions.slice(0, -1);
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
  constructor() {
    this.Name = "";
    this.Operator = Operators[0].value;
    this.Data = "";
  }
}
