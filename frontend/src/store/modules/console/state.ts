import { LogMessageType } from '@/types/console.type';

export interface ConsoleModuleState {
  messages: LogMessageType[];
  isEnabled: boolean;
}
