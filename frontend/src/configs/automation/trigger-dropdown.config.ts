import {
  AutomationActionTypes,
  NumericOperators,
} from "@/contracts/automations";
import {
  DropDownItemType,
  createDropDownItem,
  ButtonClickEventType,
} from "../../types/controls.type";

export function createNewActionDropdownItems(
  event: ButtonClickEventType
): DropDownItemType[] {
  let items: DropDownItemType[] = [];

  Object.values(AutomationActionTypes).forEach((actionType) =>
    items.push(createDropDownItem(`New ${actionType}`, actionType, event))
  );

  return items;
}

export function createStepActionOperatorsDropdowitems(
  event: ButtonClickEventType
): DropDownItemType[] {
  let items: DropDownItemType[] = [];

  Object.values(NumericOperators).forEach((operator) =>
    items.push(createDropDownItem(operator, operator, event))
  );

  return items;
}
