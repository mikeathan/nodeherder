/* eslint-disable */
declare module '*.vue' {
  import type { DefineComponent } from 'vue';
  const component: DefineComponent<{}, {}, any>;
  export default component;
}


declare namespace JSX {
  interface IntrinsicElements {
    [elem: string]: any; // Allow all elements to be JSX components
  }
}