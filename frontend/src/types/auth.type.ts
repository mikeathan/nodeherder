export type User = {
  id: string;
  username: string;
};

export type UserSession = {
  user: User;
  token: string;
  isAuthenticated: boolean;
};

export type AuthResponse = {
  status: 'ok' | 'error';
  user: User;
  token: string;
};
