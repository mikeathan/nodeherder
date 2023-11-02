export const Operators = [
  { text: "=", value: "=" },
  { text: "<=", value: "<=" },
  { text: ">=", value: ">=" },
  { text: ">", value: ">" },
  { text: "<", value: "<" },
];

export class Expose {
  constructor() {
    this.Name = null;
    this.Conditions = [];
  }

  setName(name) {
    this.Name = name;
  }
  addCondition(condition) {
    this.Conditions.push(condition);
  }
}
// "triggers": [
//   {
//     "name": "presence",
//     "conditions": [
//       {
//         "name": "presence",
//         "value": false,
//         "equality": "="
//       }
//     ],
export class Condition {
  constructor() {
    this.Name = "";
    this.Operator = Operators[0];
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
