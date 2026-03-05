import { ConsoleLevel } from '@/types/console.type';

const defaultConsoleLevelClass = 'bg-light ';
const consoleLevelClassMap = {
  error: 'bg-danger',
  critical: 'bg-danger',
  info: 'bg-info',
  warning: 'bg-warning',
  debug: 'bg-secondary',
};

export const getConsoleLevelClass = (level: ConsoleLevel) => {
  return consoleLevelClassMap[level] ?? defaultConsoleLevelClass;
};

const defaultConsoleLevelSeverity = 'secondary';
const consoleLevelSeverityMap: Record<string, string> = {
  error: 'danger',
  critical: 'danger',
  info: 'info',
  warn: 'warn',
  warning: 'warn',
  debug: 'secondary',
  trace: 'secondary',
};

export const getConsoleLevelSeverity = (level: string) => {
  return consoleLevelSeverityMap[level] ?? defaultConsoleLevelSeverity;
};
