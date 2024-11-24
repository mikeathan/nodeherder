import {
  AlertMessage,
  AlertTypes,
} from '../types/alerts.type';
import { genenerateUniqueId } from '../utils/unique';

export const createWarning = (
  message: string,
  timeout?: number,
): AlertMessage => {
  return {
    severity: AlertTypes.warning,
    title: 'Warning',
    message: message,
    id: genenerateUniqueId(),
    timeout: timeout,
    color: 'orange',
  };
};

export const createError = (
  message: string,
  timeout?: number,
): AlertMessage => {
  return {
    severity: AlertTypes.error,
    title: 'Error',
    message: message,
    id: genenerateUniqueId(),
    timeout: timeout,
    color: 'red',
  };
};

export const createSuccess = (
  message: string,
  timeout?: number,
): AlertMessage => {
  return {
    severity: AlertTypes.success,
    title: 'Success',
    message: message,
    id: genenerateUniqueId(),
    timeout: timeout,
    color: 'green',
  };
};

export const createInfo = (
  message: string,
  timeout?: number,
): AlertMessage => {
  return {
    severity: AlertTypes.info,
    title: 'Info',
    message: message,
    id: genenerateUniqueId(),
    timeout: timeout,
    color: 'blue',
  };
};
