export type DropDownItem = {
  name: string;
  value: string;
  event: (e: any) => void;
};
export type DropDownClickEvent = (e: any) => void;

export function createDropDownItem(
  name: string,
  value: string,
  event: DropDownClickEvent
): DropDownItem {
  return { name: name, value: value, event: event };
}
