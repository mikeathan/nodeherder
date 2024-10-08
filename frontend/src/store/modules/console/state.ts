import { LogMessageType } from '@/types/event-logs.type';

export interface ConsoleModuleState {
  messages: LogMessageType[];
  isEnabled: boolean;
}
