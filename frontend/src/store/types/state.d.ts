import { DeviceMap, AutomationMap } from "./types/store";

export interface State {
    devices: DeviceMap;
    automations: AutomationMap;
}
