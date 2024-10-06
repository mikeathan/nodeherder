import 'jest';
import {
  describe,
  expect,
  test,
  beforeEach,
  it,
} from '@jest/globals';
import { store } from '../../../../store/index';
import { LogMessageType } from '@/types/event-logs.type';

const mockMessage: LogMessageType = {
  level: 'info',
  message: 'test message',
  timestamp: Date.now(),
};

const mockMessages: LogMessageType[] = [
  {
    level: 'info',
    message: 'test message 1',
    timestamp: Date.now(),
  },
  {
    level: 'debug',
    message: 'test message 2',
    timestamp: Date.now(),
  },
  {
    level: 'info',
    message: 'test message 3',
    timestamp: Date.now(),
  },
  {
    level: 'critical',
    message: 'test message 4',
    timestamp: Date.now(),
  },
  {
    level: 'error',
    message: 'test message 5',
    timestamp: Date.now(),
  },
  {
    level: 'warning',
    message: 'test message 6',
    timestamp: Date.now(),
  },
];

describe('test console module', () => {
  beforeEach(() => {
    store.commit('console/clear');
  });

  it('test console messages can be added', () => {
    mockMessages.forEach((message) => {
      store.dispatch('console/addMessage', message);
    });

    store.getters['console/messages']().forEach(
      (message: LogMessageType, index: number) => {
        expect(message).toEqual(mockMessages[index]);
      },
    );
  });

  it('test console messages can be cleared', () => {
    mockMessages.forEach((message) => {
      store.dispatch('console/addMessage', message);
    });

    store.commit('console/clear');
    const messages = store.getters['console/messages']();
    expect(messages.length).toEqual(0);
  });

  it('test console messages can be deleted', () => {
    mockMessages.forEach((message) => {
      store.dispatch('console/addMessage', message);
    });

    store.commit('console/removeItems', mockMessages);
    const messages = store.getters['console/messages']();
    expect(messages.length).toEqual(0);
  });

  it('test selected console messages can be deleted', () => {
    mockMessages.forEach((message) => {
      store.dispatch('console/addMessage', message);
    });

    store.commit('console/removeItems', [
      mockMessages[0],
      mockMessages[3],
    ]);

    store.getters['console/messages']().forEach(
      (message: LogMessageType, index: number) => {
        expect(message).not.toEqual(mockMessages[0]);
        expect(message).not.toEqual(mockMessages[3]);
      },
    );
  });
});
