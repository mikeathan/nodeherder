import {
  AutomationActionTypes,
  NumericOperators,
  TriggerActionOperators,
} from "@/contracts/automations";
import {
  DropDownItemType,
  createDropDownItem,
  ButtonClickEventType,
  ButtonType,
  createButton,
  DropDownType,
  ButtonPanelType,
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

export function createTriggerActionOperatorsDropdowitems(
  event: ButtonClickEventType
): DropDownItemType[] {
  let items: DropDownItemType[] = [];

  Object.values(TriggerActionOperators).forEach((operator) =>
    items.push(createDropDownItem(operator, operator, event))
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

export function createSaveDeleteCancelButtonItems(
  saveEvent: ButtonClickEventType,
  deleteEvent: ButtonClickEventType,
  cancelEvent: ButtonClickEventType,
  isSaveDisabled?: boolean,
  isDeleteDisabled?: boolean,
  isCancelDisabled?: boolean
): ButtonType[] {
  return [
    createButton("Save", saveEvent, isSaveDisabled),
    createButton("Delete", deleteEvent, isDeleteDisabled),
    createButton("Cancel", cancelEvent, isCancelDisabled),
  ];
}

export function createStepActionButtonItems(
  dropDownItems: DropDownItemType[],
  saveEvent: ButtonClickEventType,
  deleteEvent: ButtonClickEventType,
  isSaveDisabled?: boolean,
  isDeleteDisabled?: boolean,
  isDropdownDisabled?: boolean
): ButtonPanelType[] {
  return [
    createButton("Save", saveEvent, isSaveDisabled),
    createButton("Delete", deleteEvent, isDeleteDisabled),
    createDropdown("Add Operation", dropDownItems, isDropdownDisabled),
  ];
}

export function createDropdown(
  name: string,
  items: DropDownItemType[],
  disabled: boolean = false
): DropDownType {
  return { name: name, items: items, disabled: disabled };
}
