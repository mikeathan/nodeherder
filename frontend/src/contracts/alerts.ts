import {
  AlertMessage,
  AlertType,
  AlertTypes,
} from '../types/alerts.type';
import { genenerateUniqueId } from '../utils/unique';

export const toToastSeverity = (
  severity: AlertType
): string | undefined => {
  switch (severity) {
    case 'success':
      return 'success';
    case 'error':
      return 'error';
    case 'info':
      return 'info';
    case 'warn':
      return 'warn';
    default:
      return 'info';
  }

  return 'secondary';
};

export const createWarning = (
  message: string,
  timeout?: number
): AlertMessage =>
  createAlert(
    message,
    'Warning',
    AlertTypes.warning,
    timeout
  );

export const createError = (
  message: string,
  timeout?: number
): AlertMessage =>
  createAlert(message, 'Error', AlertTypes.error, timeout);

export const createSuccess = (
  message: string,
  timeout?: number
): AlertMessage =>
  createAlert(
    message,
    'Success',
    AlertTypes.success,
    timeout
  );

export const createInfo = (
  message: string,
  timeout?: number
): AlertMessage =>
  createAlert(message, 'Info', AlertTypes.info, timeout);

const createAlert = (
  message: string,
  title: string,
  severity: AlertType,
  timeout?: number
): AlertMessage => {
  return {
    severity: severity,
    title: title,
    message: message,
    visible: false,
    id: genenerateUniqueId(),
    timeout: timeout,
  };
};
