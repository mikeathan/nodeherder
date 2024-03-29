import { AutomationActionTypes } from "@/contracts/automations";
import {
  DropDownItem,
  createDropDownItem,
  DropDownClickEvent,
} from "../../types/controls.type";

export function createNewActionDropdownItems(
  event: DropDownClickEvent
): DropDownItem[] {
  let items: DropDownItem[] = [];

  Object.values(AutomationActionTypes).forEach((actionType) =>
    items.push(createDropDownItem(`New ${actionType}`, actionType, event))
  );

  return items;
}
