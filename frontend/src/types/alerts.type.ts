import { ValueOf } from './types';

export type AlertType = ValueOf<typeof AlertTypes>;

export type AlertMessage = {
  severity: AlertType;
  title: string;
  message: string;
  id: string;
  timeout?: number;
  color: string;
};

export const AlertTypes = {
  success: 'success',
  error: 'error',
  info: 'info',
  warning: 'warning',
};
