import "jest";
import { describe, expect, test } from "@jest/globals";
import { default as automation_data1 } from "../../../../core/configs/automations/0x001788010d7d9d3f.json";
import { default as automation_data2 } from "../../../../core/configs/automations/0xa4c13894070052fc.json";

import { Automation } from "../../types/automation/automation";

test("roundtrip serializing automation data", () => {
  automationData1IsEqualToDeserializedObject();
  automationData2IsEqualToDeserializedObject();
});

function automationData1IsEqualToDeserializedObject() {
  const json = JSON.stringify(automation_data1);

  const newAutomaton: Automation = JSON.parse(json);
  isEqualToValueAndNotNull(automation_data1, "id", newAutomaton.id);
  isEqualToValueAndNotNull(
    automation_data1,
    "friendlyname",
    newAutomaton.friendlyname
  );
  isEqualToValueAndNotNull(
    automation_data1,
    "description",
    newAutomaton.description
  );
  isEqualToValueAndNotNull(automation_data1, "enabled", newAutomaton.enabled);

  automation_data1.triggers.forEach((trigger, index) => {
    const newTrigger = newAutomaton.triggers[index];
    isEqualToValueAndNotNull(trigger, "name", newTrigger.name);

    trigger.conditions.forEach((condition, contIdx) => {
      const newCondition = newTrigger.conditions[contIdx];
      isEqualToValueAndNotNull(condition, "name", newCondition.name);
      isEqualToValueAndNotNull(condition, "value", newCondition.value);
      isEqualToValueAndNotNull(condition, "equality", newCondition.equality);
    });

    const action = trigger.action;
    const newAction = newTrigger.action;

    isEqualToValueAndNotNull(action, "id", newAction.id);
    isEqualToValueAndNotNull(action, "friendlyname", newAction.friendlyname);
    isEqualToValueAndNotNull(action, "property", newAction.property);

    isEqualToValueOrNull(action, "data", newAction.data);
  });
}

function automationData2IsEqualToDeserializedObject() {
  const json = JSON.stringify(automation_data2);

  const newAutomaton: Automation = JSON.parse(json);
  isEqualToValueAndNotNull(automation_data2, "id", newAutomaton.id);
  isEqualToValueAndNotNull(
    automation_data2,
    "friendlyname",
    newAutomaton.friendlyname
  );
  isEqualToValueAndNotNull(
    automation_data2,
    "description",
    newAutomaton.description
  );
  isEqualToValueAndNotNull(automation_data2, "enabled", newAutomaton.enabled);

  automation_data2.triggers.forEach((trigger, index) => {
    const newTrigger = newAutomaton.triggers[index];
    isEqualToValueAndNotNull(trigger, "name", newTrigger.name);

    trigger.conditions.forEach((condition, contIdx) => {
      const newCondition = newTrigger.conditions[contIdx];
      isEqualToValueAndNotNull(condition, "name", newCondition.name);
      isEqualToValueAndNotNull(condition, "value", newCondition.value);
      isEqualToValueAndNotNull(condition, "equality", newCondition.equality);
    });

    const action = trigger.action;
    const newAction = newTrigger.action;

    isEqualToValueAndNotNull(action, "id", newAction.id);
    isEqualToValueAndNotNull(action, "friendlyname", newAction.friendlyname);
    isEqualToValueAndNotNull(action, "property", newAction.property);

    isEqualToValueOrNull(action, "data", newAction.data);
  });
}
function isEqualToValueAndNotNull(obj: any, propName: string, value: any) {
  expect(obj.hasOwnProperty(propName)).toBe(true);
  const srcValue = obj[propName];
  expect(value).not.toBeUndefined();
  expect(value).not.toBeNull();
  expect(value).not.toBe(null);
  expect(srcValue).toBe(value);
}

function isEqualToValueOrNull(obj: any, propName: string, value: any) {
  const srcValue = obj[propName];
  expect(srcValue).toBe(value);
}
