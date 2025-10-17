import { User } from '@/types/auth.type';

export interface AuthModuleState {
  user: User;
  token: string | null;
  authenticated: boolean;
}
