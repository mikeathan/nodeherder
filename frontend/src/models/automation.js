export const Operators = [
  { text: "=", value: "=" },
  { text: "<=", value: "<=" },
  { text: ">=", value: ">=" },
  { text: ">", value: ">" },
  { text: "<", value: "<" },
];

export class DeviceExpose {
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
}

export class ExposeCondition {
  constructor() {
    this.Name = "";
    this.Operator = Operators[0].value;
    this.Value = "";
  }
  setName(name) {
    this.Name = name;
  }
  setOperator(operator) {
    this.Operator = operator;
  }
  setValue(value) {
    this.Value = value;
  }
  Name() {
    return this.Name;
  }
  Operator() {
    return this.Operator;
  }
  Value() {
    return this.Value;
  }
}
