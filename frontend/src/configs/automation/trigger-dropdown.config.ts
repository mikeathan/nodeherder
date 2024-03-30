import {
  AutomationActionTypes,
  NumericOperators,
} from "@/contracts/automations";
import {
  DropDownItemType,
  createDropDownItem,
  ButtonClickEventType,
  ButtonType,
  createButton,
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

export function createSaveDeleteButtonItems(
  saveEvent: ButtonClickEventType,
  deleteEvent: ButtonClickEventType,
  isSaveDisabled?: boolean,
  isDeleteDisabled?: boolean
): ButtonType[] {
  return [
    createButton("Save", saveEvent, isSaveDisabled),
    createButton("Delete", deleteEvent, isDeleteDisabled),
  ];
}
