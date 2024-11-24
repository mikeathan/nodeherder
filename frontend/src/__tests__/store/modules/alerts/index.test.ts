import 'jest';
import {
  describe,
  expect,
  test,
  beforeEach,
  jest,
} from '@jest/globals';
import { store } from '../../../../store/index';
import { AlertMessage } from '../../../../types/alerts.type';

const mockAlerts: AlertMessage[] = [
  {
    severity: 'success',
    title: 'Success',
    message: 'test success message',
    timeout: 1000,
    id: '1',
    visible: false,
  },
  {
    severity: 'error',
    title: 'Error',
    message: 'test error message',
    id: '2',
    visible: false,
  },
  {
    severity: 'info',
    title: 'Info',
    message: 'test info message',
    id: '3',
    visible: false,
  },
  {
    severity: 'warning',
    title: 'Warning',
    message: 'test warning message',
    timeout: 1000,
    visible: false,
    id: '4',
  },
];

describe('test alert module', () => {
  beforeEach(() => {
    store.commit('alerts/clear');
    jest.useFakeTimers();
  });

  test('test alerts are commited to store', () => {
    Object.values(mockAlerts).forEach((message) => {
      store.commit('alerts/showAlert', message);
    });
    var allMessages = store.getters[
      'alerts/messages'
    ]() as AlertMessage[];

    Object.values(allMessages).forEach((alert, index) => {
      const message = mockAlerts[index];
      expect(alert.severity).toEqual(message.severity);
      expect(alert.title).toEqual(message.title);
      expect(alert.message).toEqual(message.message);
      expect(alert.timeout).toEqual(message.timeout);
      expect(alert.visible).toEqual(message.visible);

      expect(alert.id).toEqual(message.id);
    });
  });

  test('test alerts are cleared', () => {
    Object.values(mockAlerts).forEach((value) => {
      store.commit('alerts/showAlert', value);
    });
    var allMessages = store.getters[
      'alerts/messages'
    ]() as AlertMessage[];

    expect(allMessages.length).toEqual(4);
    store.commit('alerts/clear');

    var allMessagesAfterRemove = store.getters[
      'alerts/messages'
    ]() as AlertMessage[];

    expect(allMessagesAfterRemove.length).toEqual(0);
  });

  test('test alerts are removed using id', () => {
    Object.values(mockAlerts).forEach((value) => {
      store.commit('alerts/showAlert', value);
    });
    var allMessages = store.getters[
      'alerts/messages'
    ]() as AlertMessage[];

    expect(allMessages.length).toEqual(4);

    Object.values(mockAlerts).forEach((alert) => {
      store.commit('alerts/removeAlert', alert.id);
    });

    var allMessagesAfterRemove = store.getters[
      'alerts/messages'
    ]() as AlertMessage[];

    expect(allMessagesAfterRemove.length).toEqual(0);
  });

  test('test alerts are cleared after message timeout expires', () => {
    Object.values(mockAlerts).forEach((value) => {
      store.dispatch('alerts/showAlert', value);
    });

    jest.advanceTimersByTime(2000);

    var allMessages = store.getters[
      'alerts/messages'
    ]() as AlertMessage[];

    expect(allMessages.length).toEqual(2);
  });
});
