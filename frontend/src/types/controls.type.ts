import { KeyValuePair, ValueOf } from './types.type';

export type ControlDirection = 'horizontal' | 'vertical';
export const DashboardModes = {
  editMode: 'editmode',
  viewMode: 'viewmode',
} as const;

export type DashboardMode = ValueOf<typeof DashboardModes>;

export type ButtonPanelType = ButtonType | DropDownType;
export type ButtonType = {
  label: string;
  command: ButtonClickEventType;
  disabled: boolean;
};

export type DropDownType = {
  label: string;
  items: DropDownItemType[];
  disabled: boolean;
};

export type DropDownItemType = {
  label: string;
  value: boolean | string;
  command: ButtonClickEventType;
  icon?: string;
};

export type ButtonClickEventType = (e: any) => void;

export type TimePicker = {
  hours: number;
  minutes: number;
};

export function isDropdown(item: ButtonPanelType): boolean {
  return (item as DropDownType).items !== undefined;
}

export function createButton(name: string, click: ButtonClickEventType, disabled: boolean = false): ButtonType {
  return { label: name, command: click, disabled: disabled };
}

export function createDropDownItem(
  label: string,
  value: string | boolean,
  command: ButtonClickEventType
): DropDownItemType {
  return { label: label, value: value, command: command };
}

export function createDropdown(label: string, items: DropDownItemType[], disabled: boolean = false): DropDownType {
  return { label: label, items: items, disabled: disabled };
}

// Select
export type SelectionItems = Array<string> | KeyValuePair<string>;

export type SelectSize = keyof typeof SelectFormSize;

export const SelectFormSize = {
  normal: '',
  small: 'form-select-sm',
  large: 'form-select-lg',
} as const;

export type LayoutPosition = keyof typeof LayoutPositions;

export const LayoutPositions = {
  center: 'center',
  left: 'left',
  right: 'right',
} as const;

export type MenuBarItem = {
  to?: string;
  label?: string;
  icon?: string;
  disabled?: boolean;
  value: string | boolean;
  command: ButtonClickEventType;
  custom?: boolean;
  isLogo?: boolean;
  template?: () => void;
  children?: MenuBarItem[];
};
