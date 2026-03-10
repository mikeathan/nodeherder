import { AssistantMessage, Role } from '@/types/assistant.type';
import { genenerateUniqueId } from '@/utils/unique';

export const createUserRequest = (message: string) => {
  return createMessage('user', message);
};

export const createAssistantResponse = (message: string) => {
  return createMessage('assistant', message);
};

export const createAssistantErrorResponse = (error: Error) => {
  return createMessage('assistant', `Error: ${error.message}`);
};

const createMessage = (role: Role, message: string) :AssistantMessage=> {
  return {
    id: genenerateUniqueId(),
    role,
    content: message,
    timestamp: Date.now(),
  };
};

