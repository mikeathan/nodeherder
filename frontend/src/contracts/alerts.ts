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
    type: AlertTypes.warning,
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
    type: AlertTypes.error,
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
    type: AlertTypes.success,
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
    type: AlertTypes.info,
    message: message,
    id: genenerateUniqueId(),
    timeout: timeout,
    color: 'blue',
  };
};
