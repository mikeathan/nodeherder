import { Store } from "vuex";
import { DeviceMap } from "./types/store";

declare module "@vue/runtime-core" {
  interface State {
    devices: DeviceMap;
  }

  interface ComponentStoreProperties {
    $store: Store<State>;
  }
}
