import { User } from '@/types/auth.type';

export interface AuthModuleState {
  user: User;
  authenticated: boolean;
  validation: 'idle' | 'checking' | 'authenticated' | 'signedOut' | 'unavailable';
  revision: number;
}
