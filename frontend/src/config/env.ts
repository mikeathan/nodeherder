export const getEnv = (key: string): string => {
  const val = (import.meta as any).env?.[key] ?? (globalThis as any).process?.env?.[key];

  return val ?? '';
};
