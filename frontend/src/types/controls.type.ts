import { KeyValuePair, ValueOf } from './types.type';

export type ControlDirection = 'horizontal' | 'vertical';
export const DashboardModes= {
  editMode: 'editmode',
  viewMode: 'viewmode',
} as const;

export type DashboardMode = ValueOf<typeof DashboardModes>;

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
  value: boolean | string;
  click: ButtonClickEventType;
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
  return { name: name, click: click, disabled: disabled };
}


export function createDropDownItem(name: string, value: string | boolean, click: ButtonClickEventType): DropDownItemType {
  return { name: name, value: value, click: click };
}

export function createDropdown(name: string, items: DropDownItemType[], disabled: boolean = false): DropDownType {
  return { name: name, items: items, disabled: disabled };
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
  command?: () => void;
  custom?: boolean;
  isLogo?: boolean;
  template?: () => void;
  children?: MenuBarItem[];
};
