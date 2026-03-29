export const getEnv = (key: string, fallback: string = ''): string => {
  return process.env[key] || fallback;
};
