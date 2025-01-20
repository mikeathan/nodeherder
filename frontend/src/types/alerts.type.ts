import { ValueOf } from './types.type';

export type AlertType = ValueOf<typeof AlertTypes>;

export type AlertMessage = {
  severity: AlertType;
  title: string;
  message: string;
  id: string;
  timeout?: number;
  visible: boolean;
};

export const AlertTypes = {
  success: 'success',
  error: 'error',
  info: 'info',
  warning: 'warn',
};
