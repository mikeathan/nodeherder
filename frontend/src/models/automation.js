export const Operators = [
  { text: "=", value: "=" },
  { text: "<=", value: "<=" },
  { text: ">=", value: ">=" },
  { text: ">", value: ">" },
  { text: "<", value: "<" },
];

export class Expose {
  constructor(name) {
    this.Name = name;
    this.Conditions = [];
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
