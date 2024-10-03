import { LogMessageType } from '@/types/event-logs.type';

export interface ConsoleModuleState {
  messages: LogMessageType[]; // it ould be map with timestmap key ??
  initialized: boolean;
}
