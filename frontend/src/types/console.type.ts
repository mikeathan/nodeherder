export type LogMessageType = {
  level: ConsoleLevel;
  message: string;
  timestamp: number;
};

export type ConsolesTypes = {
  error: 'error';
  warning: 'warning';
  info: 'info';
  critical: 'critical';
  debug: 'debug';
};

export type ConsoleLevel = keyof ConsolesTypes;
