import { Store } from "vuex";
import { DeviceMap } from "./types/device";

declare module "@vue/runtime-core" {
  interface State {
    devices: DeviceMap;
  }

  interface ComponentStoreProperties {
    $store: Store<State>;
  }
}
