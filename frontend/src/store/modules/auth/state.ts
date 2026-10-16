import type { SessionValidationStatus, User } from '@/types/auth.type';

export interface AuthModuleState {
  user: User;
  authenticated: boolean;
  validation: SessionValidationStatus;
  revision: number;
}
