import { Store } from 'vuex';
import { RootState } from './state';

declare module '@vue/runtime-core' {
  interface ComponentStoreProperties {
    $store: Store<RootState>;
  }
}
