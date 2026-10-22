/*
 * @primevue/themes exposes the './aura/base' subpath at runtime, but its package type map
 * points at a file the package does not ship. Declare the shape used by the app's preset.
 */
declare module '@primevue/themes/aura/base' {
  import type { Preset } from '@primeuix/themes/types';

  export type { Preset };
  const base: Preset;
  export default base;
}
