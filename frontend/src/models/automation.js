export const Operators = ref([
  { text: "=", value: "=" },
  { text: "<=", value: "<=" },
  { text: ">=", value: ">=" },
  { text: ">", value: ">" },
  { text: "<", value: "<" },
]);

export default class Condition {
  Name = null;
  Operator = "";
  Value = null;

  constructor(name, operator, value) {
    this.Name = name;
    this.Operator = operator;
    this.value = value;
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
