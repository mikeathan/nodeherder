/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_WS_BASE_URL: string;
  // Add other variables here as you create them
  readonly VITE_API_URL?: string;
  readonly VITE_DEBUG_MODE?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}