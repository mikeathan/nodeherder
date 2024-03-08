interface operationItem {
  name: string;
  value: number;
}
type operationitems = Array<operationItem>;

const ActionOperationsMap: { [K in string]: operationItem } = {
  none: { name: 'NoOp', value: 0 },
  step_increase: { name: 'Step increase', value: 1 },
  step_decrease: { name: 'Step decrease', value: 2 },
  presets: { name: 'Rotation', value: 3 }
} as const;

export function resolveStepOperations() {
  const arr = Array<operationItem>(
    ActionOperationsMap.step_increase,
    ActionOperationsMap.step_decrease
  );
  return Object.assign({}, ...arr.map(x => ({ [x.name]: x.value })));
}

export function resolveObjectOperations(obj: any) {
  const arr: operationitems = [ActionOperationsMap.none];
  if (obj.hasOwnProperty('type')) {
    if (obj['type'] === 'numeric') {
      arr.push(ActionOperationsMap.step_increase);
      arr.push(ActionOperationsMap.step_decrease);
    }
  }
  if (obj.hasOwnProperty('presets')) {
    if (obj['presets'] !== undefined) {
      arr.push(ActionOperationsMap.presets);
    }
  }

  return Object.assign({}, ...arr.map(x => ({ [x.name]: x.value })));
}

export enum OperationType {
  NoOperation = 0,
  StepOperation = -1,
  RotationOperation = 3,
  StepIncreaseOperation = 1,
  StepDecreaseOperation = 2
}
