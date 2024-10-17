import 'jest';
import {
  describe,
  expect,
  test,
  beforeEach,
  it,
} from '@jest/globals';
import { store } from '../../../../store/index';
import { LogMessageType } from '@/types/console.type';

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

  it('test expired console messages are removed', () => {
    let mockMessages: LogMessageType[] = [];

    // create 10 messages with timestamps 1  minutes back
    const now = new Date();
    for (let i = 0; i < 10; i++) {
      mockMessages.push({
        level: 'info',
        message: 'test message ' + i,
        timestamp: new Date(
          now.getTime() - i * 61000,
        ).getTime(),
      });
    }

    mockMessages.forEach((message) => {
      store.dispatch('console/addMessage', message);
    });

    // remove messages older than 5 minutes
    store.commit(
      'console/removeExpiredMessages',
      5 * 60000,
    );

    const messages = store.getters[
      'console/messages'
    ]() as LogMessageType[];

    // expect 5 messages to be left
    expect(messages.length).toEqual(5);
  });
});
