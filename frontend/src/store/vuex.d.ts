import { Store } from "vuex";
import { RootState } from "./types/state";

declare module "@vue/runtime-core" {
  interface ComponentStoreProperties {
    $store: Store<RootState>;
  }
}
