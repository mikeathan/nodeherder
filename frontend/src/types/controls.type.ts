export type ButtonPanelType = ButtonType | DropDownType;
export type ButtonType = {
  name: string;
  click: ButtonClickEventType;
  disabled: boolean;
};

export type DropDownType = {
  name: string;
  items: DropDownItemType[];
  disabled: boolean;
};

export type DropDownItemType = {
  name: string;
  value: string;
  event: (e: any) => void;
};

export type ButtonClickEventType = (e: any) => void;

export function isButton(item: ButtonPanelType): item is ButtonType {
  return (item as ButtonType) !== undefined;
}

export function isDropdown(item: ButtonPanelType): item is DropDownType {
  return (item as DropDownType) !== undefined;
}

export function createButton(
  name: string,
  click: ButtonClickEventType,
  disabled: boolean = false
): ButtonType {
  return { name: name, click: click, disabled: disabled };
}

export function createDropDownItem(
  name: string,
  value: string,
  click: ButtonClickEventType
): DropDownItemType {
  return { name: name, value: value, event: click };
}

export function createDropdown(
  name: string,
  items: DropDownItemType[],
  disabled: boolean = false
): DropDownType {
  return { name: name, items: items, disabled: disabled };
}
